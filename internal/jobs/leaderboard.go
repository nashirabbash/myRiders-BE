package jobs

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/robfig/cron/v3"
	"github.com/nashirabbash/trackride/internal/db"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
)

// LeaderboardJob handles periodic leaderboard ranking computation
type LeaderboardJob struct {
	queries  *db.Queries
	cron     *cron.Cron
	timezone string
}

// NewLeaderboardJob creates a new leaderboard job.
// timezone should come from config (e.g. "Asia/Jakarta").
func NewLeaderboardJob(queries *db.Queries, timezone string) *LeaderboardJob {
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}
	return &LeaderboardJob{
		queries:  queries,
		cron:     cron.New(cron.WithLocation(mustLoadLocation(timezone))),
		timezone: timezone,
	}
}

// Start begins the cron job scheduling
func (j *LeaderboardJob) Start() {
	// Schedule weekly computation every Monday at 00:01 WIB
	_, err := j.cron.AddFunc("1 0 * * MON", j.computeWeekly)
	if err != nil {
		log.Printf("[Leaderboard] Failed to add cron job: %v", err)
		return
	}

	j.cron.Start()
	log.Println("[Leaderboard] Cron job started (weekly on Monday 00:01 WIB)")
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

	// 1. Delete old entries for this period to avoid duplicates
	deleteParams := sqlc.DeleteLeaderboardEntriesParams{
		PeriodType:  "weekly",
		PeriodStart: pgTypeDate(periodStart),
		VehicleType: sqlc.NullVehicleType{}, // NULL for all vehicles
	}
	if err := j.queries.DeleteLeaderboardEntries(ctx, deleteParams); err != nil {
		log.Printf("[Leaderboard] Error deleting old entries: %v", err)
		return
	}
	log.Printf("[Leaderboard] Deleted old entries for weekly period starting %s", periodStart.Format("2006-01-02"))

	// 2. Compute and insert rankings for all vehicles combined
	periodStartTS := pgTypeTimestamptz(periodStart)
	allRankings, err := j.queries.ComputeWeeklyRankings(ctx, periodStartTS)
	if err != nil {
		log.Printf("[Leaderboard] Error computing weekly rankings: %v", err)
		return
	}

	for rank, row := range allRankings {
		insertParams := sqlc.InsertLeaderboardEntryParams{
			UserID:      row.UserID,
			VehicleType: sqlc.NullVehicleType{}, // NULL for all vehicles
			PeriodType:  "weekly",
			PeriodStart: pgTypeDate(periodStart),
			TotalKm:     row.TotalKm,
			TotalRides:  int32(row.TotalRides),
			Rank:        int32(rank + 1),
		}
		if err := j.queries.InsertLeaderboardEntry(ctx, insertParams); err != nil {
			log.Printf("[Leaderboard] Error inserting all-vehicle ranking for user %s: %v", row.UserID, err)
			continue
		}
	}
	log.Printf("[Leaderboard] Inserted %d all-vehicle rankings", len(allRankings))

	// 3. Compute and insert rankings for each vehicle type
	vehicleTypes := []string{"motor", "mobil", "sepeda"}
	for _, vehicleType := range vehicleTypes {
		vehicleRankings, err := j.queries.ComputeWeeklyRankingsByVehicle(ctx, sqlc.ComputeWeeklyRankingsByVehicleParams{
			StartedAt: periodStartTS,
			Type:      vehicleType,
		})
		if err != nil {
			log.Printf("[Leaderboard] Error computing rankings for vehicle type %s: %v", vehicleType, err)
			continue
		}

		for rank, row := range vehicleRankings {
			insertParams := sqlc.InsertLeaderboardEntryParams{
				UserID:      row.UserID,
				VehicleType: sqlc.NullVehicleType{VehicleType: sqlc.VehicleType(vehicleType), Valid: true},
				PeriodType:  "weekly",
				PeriodStart: pgTypeDate(periodStart),
				TotalKm:     row.TotalKm,
				TotalRides:  int32(row.TotalRides),
				Rank:        int32(rank + 1),
			}
			if err := j.queries.InsertLeaderboardEntry(ctx, insertParams); err != nil {
				log.Printf("[Leaderboard] Error inserting %s ranking for user %s: %v", vehicleType, row.UserID, err)
				continue
			}
		}
		log.Printf("[Leaderboard] Inserted %d %s vehicle rankings", len(vehicleRankings), vehicleType)
	}

	log.Printf("[Leaderboard] Weekly computation complete for %s", periodStart.Format("2006-01-02"))
}

// getLastMonday returns the date of the last Monday in the given timezone
func getLastMonday(timezone string) time.Time {
	now := time.Now().In(mustLoadLocation(timezone))
	daysBack := int(now.Weekday()) - int(time.Monday)
	if daysBack < 0 {
		daysBack += 7
	}
	d := now.AddDate(0, 0, -daysBack)
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

// pgTypeDate converts a time.Time to pgtype.Date
func pgTypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

// pgTypeTimestamptz converts a time.Time to pgtype.Timestamptz
func pgTypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t.UTC(),
		Valid: true,
	}
}
