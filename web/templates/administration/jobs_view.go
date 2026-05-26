package adminviews

import (
	"fmt"
	"time"

	"dorm-man/internal/models"
)

func jobPriorityClass(priority models.JobPriority) string {
	switch priority {
	case models.JobPriorityHigh:
		return "dm-job-badge-priority-high"
	case models.JobPriorityLow:
		return "dm-job-badge-priority-low"
	default:
		return "dm-job-badge-priority-medium"
	}
}

func jobStatusClass(status models.JobStatus) string {
	switch status {
	case models.JobStatusInProgress:
		return "dm-job-badge-status-in-progress"
	case models.JobStatusFinished:
		return "dm-job-badge-status-finished"
	case models.JobStatusCanceled:
		return "dm-job-badge-status-canceled"
	case models.JobStatusScheduled:
		return "dm-job-badge-status-scheduled"
	default:
		return "dm-job-badge-status-planned"
	}
}

func jobDueLabel(endsAt time.Time) string {
	return fmt.Sprintf("Due %s", endsAt.Format("2006-01-02 15:04"))
}
