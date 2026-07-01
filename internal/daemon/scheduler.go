package daemon

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"codeforge/internal/logger"

	"github.com/robfig/cron/v3"
)

// Scheduler coordinates timed trigger actions using robfig/cron.
type Scheduler struct {
	cronClient *cron.Cron
	daemon     *Daemon
	logger     *logger.Logger
	cronIDs    map[string][]cron.EntryID // project -> list of cron entry IDs
}

// NewScheduler creates a Scheduler instance.
func NewScheduler(d *Daemon, l *logger.Logger) *Scheduler {
	return &Scheduler{
		cronClient: cron.New(cron.WithSeconds()), // Standard cron with seconds precision optionally or standard 5-field cron. cron.New() defaults to 5-field
		daemon:     d,
		logger:     l,
		cronIDs:    make(map[string][]cron.EntryID),
	}
}

// Start initiates the cron schedule thread.
func (s *Scheduler) Start(ctx context.Context) {
	s.cronClient.Start()
	go func() {
		<-ctx.Done()
		s.cronClient.Stop()
	}()
}

// RegisterPipelineSchedules parses triggers and sets up cron jobs.
func (s *Scheduler) RegisterPipelineSchedules(project string, cronExprs []string) {
	s.UnregisterPipelineSchedules(project)

	var ids []cron.EntryID
	for _, rawExpr := range cronExprs {
		cronExpr := parseCronShorthand(rawExpr)
		id, err := s.cronClient.AddFunc(cronExpr, func() {
			s.logger.Log(project, "INFO", "Cron timer fired (%s)", rawExpr)
			s.daemon.Trigger(project, "Scheduler trigger")
		})
		if err != nil {
			s.logger.Log(project, "ERROR", "Failed to add cron schedule %q: %v", rawExpr, err)
			continue
		}
		ids = append(ids, id)
	}

	s.cronIDs[project] = ids
}

// UnregisterPipelineSchedules cancels registered schedules for a project.
func (s *Scheduler) UnregisterPipelineSchedules(project string) {
	if ids, ok := s.cronIDs[project]; ok {
		for _, id := range ids {
			s.cronClient.Remove(id)
		}
		delete(s.cronIDs, project)
	}
}

var shorthandMinRegex = regexp.MustCompile(`^(\d+)m$`)
var shorthandHourRegex = regexp.MustCompile(`^(\d+)h$`)

func parseCronShorthand(expr string) string {
	expr = strings.TrimSpace(strings.ToLower(expr))

	// check "5m" -> "*/5 * * * *"
	if matches := shorthandMinRegex.FindStringSubmatch(expr); len(matches) == 2 {
		return fmt.Sprintf("*/%s * * * *", matches[1])
	}
	// check "2h" -> "0 */2 * * *"
	if matches := shorthandHourRegex.FindStringSubmatch(expr); len(matches) == 2 {
		return fmt.Sprintf("0 */%s * * *", matches[1])
	}

	switch expr {
	case "hourly":
		return "0 * * * *"
	case "daily":
		return "0 0 * * *"
	case "weekly":
		return "0 0 * * 0"
	case "monthly":
		return "0 0 1 * *"
	}

	return expr
}
