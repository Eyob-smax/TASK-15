// Package jobs provides background job runners and scheduled task implementations
// for the FitCommerce platform.
//
// This package will contain the following job implementations (Prompts 4 and 7):
//
//   - Auto-close job: periodically scans for orders in "created" status that have
//     exceeded the 30-minute auto-close timeout and transitions them to "auto_closed".
//
//   - Cutoff evaluation job: checks active group-buy campaigns at their cutoff time,
//     evaluating whether the minimum quantity threshold has been met, and transitions
//     campaigns to "succeeded" or "failed" accordingly.
//
//   - Backup job: orchestrates database backup creation, compression, SHA-256 checksum
//     computation, encryption of the archive, and recording of the backup run metadata.
//
//   - Retention cleanup job: identifies entities that have exceeded their retention
//     policy period and performs soft-delete or archival as appropriate, respecting
//     the 7-year financial and 2-year access log retention requirements.
//
//   - Variance deadline job: scans open variance records that are past their
//     resolution due date and triggers escalation notifications for overdue items.
//
// All jobs are designed to be run as goroutines with configurable intervals and
// graceful shutdown support via context cancellation.
package jobs

import (
	"context"
	"log/slog"
	"time"

	"fitcommerce/internal/application"
)

// AutoCloseJob periodically closes orders that have exceeded the unpaid timeout.
type AutoCloseJob struct {
	orderSvc application.OrderService
	logger   *slog.Logger
	interval time.Duration
}

// NewAutoCloseJob creates an AutoCloseJob with a 1-minute default interval.
func NewAutoCloseJob(svc application.OrderService, logger *slog.Logger) *AutoCloseJob {
	return &AutoCloseJob{
		orderSvc: svc,
		logger:   logger,
		interval: time.Minute,
	}
}

// NewAutoCloseJobWithInterval creates an AutoCloseJob with the given interval.
// Use this constructor in tests to reduce polling frequency.
func NewAutoCloseJobWithInterval(svc application.OrderService, logger *slog.Logger, interval time.Duration) *AutoCloseJob {
	return &AutoCloseJob{
		orderSvc: svc,
		logger:   logger,
		interval: interval,
	}
}

// Run starts the auto-close polling loop. It exits when ctx is cancelled.
func (j *AutoCloseJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			count, err := j.orderSvc.AutoCloseExpired(ctx, t)
			if err != nil {
				j.logger.Error("auto-close error", "error", err)
				continue
			}
			if count > 0 {
				j.logger.Info("auto-closed orders", "count", count)
			}
		}
	}
}

// CutoffEvalJob periodically evaluates group-buy campaigns that have passed their cutoff.
type CutoffEvalJob struct {
	campaignSvc application.CampaignService
	logger      *slog.Logger
	interval    time.Duration
}

// NewCutoffEvalJob creates a CutoffEvalJob with a 1-minute default interval.
func NewCutoffEvalJob(svc application.CampaignService, logger *slog.Logger) *CutoffEvalJob {
	return &CutoffEvalJob{
		campaignSvc: svc,
		logger:      logger,
		interval:    time.Minute,
	}
}

// NewCutoffEvalJobWithInterval creates a CutoffEvalJob with the given interval.
// Use this constructor in tests to reduce polling frequency.
func NewCutoffEvalJobWithInterval(svc application.CampaignService, logger *slog.Logger, interval time.Duration) *CutoffEvalJob {
	return &CutoffEvalJob{
		campaignSvc: svc,
		logger:      logger,
		interval:    interval,
	}
}

// Run starts the cutoff evaluation polling loop. It exits when ctx is cancelled.
func (j *CutoffEvalJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			campaigns, err := j.campaignSvc.ListPastCutoff(ctx, t)
			if err != nil {
				j.logger.Error("cutoff-eval list error", "error", err)
				continue
			}
			for _, c := range campaigns {
				if err := j.campaignSvc.EvaluateAtCutoff(ctx, c.ID, t); err != nil {
					j.logger.Error("cutoff-eval error", "campaign_id", c.ID, "error", err)
				}
			}
		}
	}
}
