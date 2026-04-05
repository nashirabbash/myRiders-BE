package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/robfig/cron/v3"
	"github.com/nashirabbash/trackride/internal/db"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
)

// LeaderboardJob handles periodic leaderboard ranking computation
type LeaderboardJob struct {
	queries  *db.Queries
	pool     *pgxpool.Pool
	cron     *cron.Cron
	timezone string
}

// NewLeaderboardJob creates a new leaderboard job.
// pool and timezone should come from main application setup.
func NewLeaderboardJob(queries *db.Queries, pool *pgxpool.Pool, timezone string) *LeaderboardJob {
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}
	return &LeaderboardJob{
		queries:  queries,
		pool:     pool,
		cron:     cron.New(cron.WithLocation(mustLoadLocation(timezone))),
		timezone: timezone,
	}
}

// Start begins the cron job scheduling
func (j *LeaderboardJob) Start() {
	// Schedule weekly computation every Monday at 00:01 WIB
	_, err := j.cron.AddFunc("1 0 * * MON", j.computeWeekly)
	if err != nil {
		log.Printf("[Leaderboard] Failed to add weekly cron job: %v", err)
		return
	}

	// Schedule monthly computation first day of month at 00:02 WIB
	_, err = j.cron.AddFunc("2 0 1 * *", j.computeMonthly)
	if err != nil {
		log.Printf("[Leaderboard] Failed to add monthly cron job: %v", err)
		return
	}

	j.cron.Start()
	log.Println("[Leaderboard] Cron jobs started (weekly on Monday 00:01 WIB, monthly on 1st at 00:02 WIB)")
}

// Stop stops the cron job scheduler and returns a context that signals when
// all in-flight jobs have completed
func (j *LeaderboardJob) Stop() context.Context {
	return j.cron.Stop()
}

// computeWeekly computes weekly leaderboard rankings
func (j *LeaderboardJob) computeWeekly() {
	ctx := context.Background()
	periodStart := getLastMonday(j.timezone)

	log.Printf("[Leaderboard] Computing weekly rankings for period starting %s", periodStart.Format("2006-01-02"))

	if err := j.computeWeeklyRankings(ctx, periodStart); err != nil {
		log.Printf("[Leaderboard] Error computing weekly rankings: %v", err)
		return
	}

	log.Printf("[Leaderboard] Weekly computation complete for %s", periodStart.Format("2006-01-02"))
}

// computeMonthly computes monthly leaderboard rankings
func (j *LeaderboardJob) computeMonthly() {
	ctx := context.Background()
	now := time.Now().In(mustLoadLocation(j.timezone))
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	log.Printf("[Leaderboard] Computing monthly rankings for period starting %s", periodStart.Format("2006-01-02"))

	if err := j.computeMonthlyRankings(ctx, periodStart); err != nil {
		log.Printf("[Leaderboard] Error computing monthly rankings: %v", err)
		return
	}

	log.Printf("[Leaderboard] Monthly computation complete for %s", periodStart.Format("2006-01-02"))
}

// computeWeeklyRankings computes and inserts weekly rankings in a transaction
func (j *LeaderboardJob) computeWeeklyRankings(ctx context.Context, periodStart time.Time) error {
	// Start a transaction to ensure atomicity
	tx, err := j.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := j.queries.WithTx(tx)
	periodStartDate := pgTypeDate(periodStart)
	periodStartTS := pgTypeTimestamptz(periodStart)

	// Delete old entries (both all-vehicle and vehicle-specific)
	if err := deleteLeaderboardPeriod(ctx, txQueries, "weekly", periodStartDate); err != nil {
		return err
	}

	// Compute and insert all-vehicle rankings
	allRankings, err := txQueries.ComputeWeeklyRankings(ctx, periodStartTS)
	if err != nil {
		return fmt.Errorf("failed to compute weekly rankings: %w", err)
	}

	// TODO: Optimize N+1 INSERT queries by using bulk insert with sqlc :copyfrom or unnesting arrays
	for rank, row := range allRankings {
		insertParams := sqlc.InsertLeaderboardEntryParams{
			UserID:      row.UserID,
			VehicleType: sqlc.NullVehicleType{}, // NULL for all vehicles
			PeriodType:  "weekly",
			PeriodStart: periodStartDate,
			TotalKm:     row.TotalKm,
			TotalRides:  int32(row.TotalRides),
			Rank:        int32(rank + 1),
		}
		if err := txQueries.InsertLeaderboardEntry(ctx, insertParams); err != nil {
			return fmt.Errorf("failed to insert all-vehicle ranking: %w", err)
		}
	}
	log.Printf("[Leaderboard] Inserted %d all-vehicle weekly rankings", len(allRankings))

	// Compute and insert vehicle-specific rankings
	vehicleTypes := []sqlc.VehicleType{"motor", "mobil", "sepeda"}
	for _, vehicleType := range vehicleTypes {
		vehicleRankings, err := txQueries.ComputeWeeklyRankingsByVehicle(ctx, sqlc.ComputeWeeklyRankingsByVehicleParams{
			StartedAt: periodStartTS,
			Type:      vehicleType,
		})
		if err != nil {
			return fmt.Errorf("failed to compute weekly rankings for %s: %w", vehicleType, err)
		}

		for rank, row := range vehicleRankings {
			insertParams := sqlc.InsertLeaderboardEntryParams{
				UserID:      row.UserID,
				VehicleType: sqlc.NullVehicleType{VehicleType: vehicleType, Valid: true},
				PeriodType:  "weekly",
				PeriodStart: periodStartDate,
				TotalKm:     row.TotalKm,
				TotalRides:  int32(row.TotalRides),
				Rank:        int32(rank + 1),
			}
			if err := txQueries.InsertLeaderboardEntry(ctx, insertParams); err != nil {
				return fmt.Errorf("failed to insert %s weekly ranking: %w", vehicleType, err)
			}
		}
		log.Printf("[Leaderboard] Inserted %d %s weekly rankings", len(vehicleRankings), vehicleType)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit weekly transaction: %w", err)
	}
	return nil
}

// computeMonthlyRankings computes and inserts monthly rankings in a transaction
func (j *LeaderboardJob) computeMonthlyRankings(ctx context.Context, periodStart time.Time) error {
	// Start a transaction to ensure atomicity
	tx, err := j.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := j.queries.WithTx(tx)
	periodStartDate := pgTypeDate(periodStart)
	periodStartTS := pgTypeTimestamptz(periodStart)

	// Delete old entries (both all-vehicle and vehicle-specific)
	if err := deleteLeaderboardPeriod(ctx, txQueries, "monthly", periodStartDate); err != nil {
		return err
	}

	// Compute and insert all-vehicle rankings
	allRankings, err := txQueries.ComputeMonthlyRankings(ctx, periodStartTS)
	if err != nil {
		return fmt.Errorf("failed to compute monthly rankings: %w", err)
	}

	// TODO: Optimize N+1 INSERT queries by using bulk insert with sqlc :copyfrom or unnesting arrays
	for rank, row := range allRankings {
		insertParams := sqlc.InsertLeaderboardEntryParams{
			UserID:      row.UserID,
			VehicleType: sqlc.NullVehicleType{}, // NULL for all vehicles
			PeriodType:  "monthly",
			PeriodStart: periodStartDate,
			TotalKm:     row.TotalKm,
			TotalRides:  int32(row.TotalRides),
			Rank:        int32(rank + 1),
		}
		if err := txQueries.InsertLeaderboardEntry(ctx, insertParams); err != nil {
			return fmt.Errorf("failed to insert all-vehicle ranking: %w", err)
		}
	}
	log.Printf("[Leaderboard] Inserted %d all-vehicle monthly rankings", len(allRankings))

	// Compute and insert vehicle-specific rankings
	vehicleTypes := []sqlc.VehicleType{"motor", "mobil", "sepeda"}
	for _, vehicleType := range vehicleTypes {
		vehicleRankings, err := txQueries.ComputeMonthlyRankingsByVehicle(ctx, sqlc.ComputeMonthlyRankingsByVehicleParams{
			StartedAt: periodStartTS,
			Type:      vehicleType,
		})
		if err != nil {
			return fmt.Errorf("failed to compute monthly rankings for %s: %w", vehicleType, err)
		}

		for rank, row := range vehicleRankings {
			insertParams := sqlc.InsertLeaderboardEntryParams{
				UserID:      row.UserID,
				VehicleType: sqlc.NullVehicleType{VehicleType: vehicleType, Valid: true},
				PeriodType:  "monthly",
				PeriodStart: periodStartDate,
				TotalKm:     row.TotalKm,
				TotalRides:  int32(row.TotalRides),
				Rank:        int32(rank + 1),
			}
			if err := txQueries.InsertLeaderboardEntry(ctx, insertParams); err != nil {
				return fmt.Errorf("failed to insert %s monthly ranking: %w", vehicleType, err)
			}
		}
		log.Printf("[Leaderboard] Inserted %d %s monthly rankings", len(vehicleRankings), vehicleType)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit monthly transaction: %w", err)
	}
	return nil
}

// deleteLeaderboardPeriod deletes all leaderboard entries for a given period
func deleteLeaderboardPeriod(ctx context.Context, txQueries *sqlc.Queries, periodType string, periodStartDate pgtype.Date) error {
	// Delete all-vehicle entries
	deleteParams := sqlc.DeleteLeaderboardEntriesParams{
		PeriodType:  periodType,
		PeriodStart: periodStartDate,
		VehicleType: sqlc.NullVehicleType{}, // NULL for all vehicles
	}
	if err := txQueries.DeleteLeaderboardEntries(ctx, deleteParams); err != nil {
		return fmt.Errorf("failed to delete old all-vehicle entries: %w", err)
	}

	// Delete vehicle-specific entries
	vehicleTypes := []sqlc.VehicleType{"motor", "mobil", "sepeda"}
	for _, vehicleType := range vehicleTypes {
		deleteParams.VehicleType = sqlc.NullVehicleType{VehicleType: vehicleType, Valid: true}
		if err := txQueries.DeleteLeaderboardEntries(ctx, deleteParams); err != nil {
			return fmt.Errorf("failed to delete old %s entries: %w", vehicleType, err)
		}
	}

	log.Printf("[Leaderboard] Deleted old entries for %s period starting %s", periodType, periodStartDate.Time.Format("2006-01-02"))
	return nil
}

// getLastMonday returns the date of the last Monday in the given timezone, explicitly truncated to midnight
func getLastMonday(timezone string) time.Time {
	now := time.Now().In(mustLoadLocation(timezone))
	daysBack := int(now.Weekday()) - int(time.Monday)
	if daysBack < 0 {
		daysBack += 7
	}
	d := now.AddDate(0, 0, -daysBack)
	// Explicitly truncate to midnight (00:00:00)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// mustLoadLocation loads a timezone location, defaulting to UTC on error
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// pgTypeDate converts a time.Time to pgtype.Date, explicitly truncated to midnight
func pgTypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()),
		Valid: true,
	}
}

// pgTypeTimestamptz converts a time.Time to pgtype.Timestamptz in UTC
func pgTypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t.UTC(),
		Valid: true,
	}
}
