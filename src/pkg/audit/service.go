package audit

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Service struct {
	repo      Repository
	enabled   bool
	retention time.Duration
}

type Config struct {
	Enabled   bool
	Retention time.Duration
}

func NewService(repo Repository, config *Config) *Service {
	if config == nil {
		config = &Config{
			Enabled:   true,
			Retention: 90 * 24 * time.Hour, // 90 days default
		}
	}
	
	return &Service{
		repo:      repo,
		enabled:   config.Enabled,
		retention: config.Retention,
	}
}

func (s *Service) Log(ctx context.Context, event *AuditEvent) error {
	if !s.enabled {
		return nil
	}
	
	return s.repo.Log(ctx, event)
}

func (s *Service) LogAction(ctx context.Context, userID int64, username, action, resource string, resourceID int64, details string, success bool, err error) error {
	event := &AuditEvent{
		Timestamp:  time.Now(),
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		Success:    success,
	}
	
	if err != nil {
		event.Error = err.Error()
	}
	
	return s.Log(ctx, event)
}

func (s *Service) Query(ctx context.Context, filter *AuditFilter) ([]AuditEvent, error) {
	if !s.enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}
	
	return s.repo.Query(ctx, filter)
}

func (s *Service) Count(ctx context.Context, filter *AuditFilter) (int64, error) {
	if !s.enabled {
		return 0, fmt.Errorf("audit logging is disabled")
	}
	
	return s.repo.Count(ctx, filter)
}

func (s *Service) Cleanup(ctx context.Context) (int64, error) {
	if !s.enabled {
		return 0, nil
	}
	
	olderThan := time.Now().Add(-s.retention)
	count, err := s.repo.Cleanup(ctx, olderThan)
	if err != nil {
		return 0, err
	}
	
	if count > 0 {
		log.Printf("Cleaned up %d audit log entries older than %s", count, olderThan.Format(time.RFC3339))
	}
	
	return count, nil
}

func (s *Service) StartCleanupScheduler(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := s.Cleanup(ctx)
			if err != nil {
				log.Printf("Audit cleanup error: %v", err)
			}
		}
	}
}

func (s *Service) IsEnabled() bool {
	return s.enabled
}

func (s *Service) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *Service) GetRetention() time.Duration {
	return s.retention
}

func (s *Service) SetRetention(retention time.Duration) {
	s.retention = retention
}
