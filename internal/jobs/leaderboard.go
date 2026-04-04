package jobs

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
)

// LeaderboardJob handles periodic leaderboard ranking computation
type LeaderboardJob struct {
	queries *sqlc.Queries
	cron    *cron.Cron
}

// NewLeaderboardJob creates a new leaderboard job
func NewLeaderboardJob(queries *sqlc.Queries) *LeaderboardJob {
	return &LeaderboardJob{
		queries: queries,
		cron:    cron.New(cron.WithLocation(mustLoadLocation("Asia/Jakarta"))),
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

// Stop stops the cron job scheduler
func (j *LeaderboardJob) Stop() {
	j.cron.Stop()
	log.Println("[Leaderboard] Cron job stopped")
}

// computeWeekly computes weekly leaderboard rankings
func (j *LeaderboardJob) computeWeekly() {
	ctx := context.Background()
	periodStart := getLastMonday()

	log.Printf("[Leaderboard] Computing weekly rankings for period starting %s", periodStart.Format("2006-01-02"))

	// TODO: Implement actual leaderboard computation
	// This will involve:
	// 1. Deleting old entries for this period
	// 2. Computing rankings by total_km for each vehicle type
	// 3. Inserting new entries with ranks

	log.Printf("[Leaderboard] Weekly computation complete for %s", periodStart.Format("2006-01-02"))
	_ = ctx
}

// getLastMonday returns the date of the last Monday
func getLastMonday() time.Time {
	now := time.Now().In(mustLoadLocation("Asia/Jakarta"))
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
