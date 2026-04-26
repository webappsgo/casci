package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"text/template"
	"time"
)

// Service manages notification sending
type Service struct {
	repo      Repository
	senders   map[NotificationType]Sender
	templates *template.Template
	queue     chan *NotificationRequest
	mu        sync.RWMutex
}

// NotificationRequest represents a queued notification
type NotificationRequest struct {
	Config *NotificationConfig
	Data   *NotificationData
	Result chan *NotificationResult
}

// Repository interface for notification configuration persistence
type Repository interface {
	CreateConfig(ctx context.Context, config *NotificationConfig) error
	GetConfig(ctx context.Context, id int) (*NotificationConfig, error)
	GetConfigsByUser(ctx context.Context, userID int) ([]*NotificationConfig, error)
	GetConfigsByProject(ctx context.Context, projectID int) ([]*NotificationConfig, error)
	UpdateConfig(ctx context.Context, config *NotificationConfig) error
	DeleteConfig(ctx context.Context, id int) error
	LogNotification(ctx context.Context, log *NotificationLog) error
	GetNotificationLogs(ctx context.Context, buildID int) ([]*NotificationLog, error)
}

// Sender interface for notification channel implementations
type Sender interface {
	Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error
	Type() NotificationType
}

// NewService creates a new notification service
func NewService(repo Repository) *Service {
	s := &Service{
		repo:    repo,
		senders: make(map[NotificationType]Sender),
		queue:   make(chan *NotificationRequest, 1000),
	}

	// Initialize default templates
	s.initTemplates()

	// Register default senders
	s.RegisterSender(NewSlackSender())
	s.RegisterSender(NewDiscordSender())
	s.RegisterSender(NewEmailSender())
	s.RegisterSender(NewWebhookSender())
	s.RegisterSender(NewGitHubStatusSender())
	s.RegisterSender(NewGitLabStatusSender())

	// Start worker pool
	for i := 0; i < 10; i++ {
		go s.worker()
	}

	return s
}

// initTemplates initializes template engine
func (s *Service) initTemplates() {
	s.templates = template.New("notifications")

	// Parse default templates
	template.Must(s.templates.New("slack").Parse(DefaultSlackTemplate))
	template.Must(s.templates.New("email_subject").Parse(DefaultEmailSubject))
	template.Must(s.templates.New("email_body").Parse(DefaultEmailBody))
	template.Must(s.templates.New("webhook").Parse(DefaultWebhookTemplate))
}

// RegisterSender registers a notification sender
func (s *Service) RegisterSender(sender Sender) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.senders[sender.Type()] = sender
	log.Printf("Registered notification sender: %s", sender.Type())
}

// SendNotification sends a notification synchronously
func (s *Service) SendNotification(ctx context.Context, config *NotificationConfig, data *NotificationData) *NotificationResult {
	start := time.Now()
	result := &NotificationResult{
		ConfigID: config.ID,
		SentAt:   start,
	}

	// Check if enabled
	if !config.Enabled {
		result.Success = false
		result.Error = "notification disabled"
		result.Duration = time.Since(start)
		return result
	}

	// Check if event matches
	if !s.matchesEvent(config.Events, data.Event) {
		result.Success = false
		result.Error = "event not in configured events"
		result.Duration = time.Since(start)
		return result
	}

	// Evaluate filter if present
	if config.Filter != "" {
		matches, err := s.evaluateFilter(config.Filter, data)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("filter evaluation error: %v", err)
			result.Duration = time.Since(start)
			return result
		}
		if !matches {
			result.Success = false
			result.Error = "filter did not match"
			result.Duration = time.Since(start)
			return result
		}
	}

	// Get sender
	s.mu.RLock()
	sender, exists := s.senders[config.Type]
	s.mu.RUnlock()

	if !exists {
		result.Success = false
		result.Error = fmt.Sprintf("no sender for type: %s", config.Type)
		result.Duration = time.Since(start)
		return result
	}

	// Send notification
	if err := sender.Send(ctx, config, data); err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	result.Success = true
	result.Duration = time.Since(start)

	// Log notification
	go s.logNotification(config.ID, data.BuildID, data.Event, result)

	return result
}

// SendNotificationAsync queues a notification for asynchronous sending
func (s *Service) SendNotificationAsync(config *NotificationConfig, data *NotificationData) {
	req := &NotificationRequest{
		Config: config,
		Data:   data,
		Result: make(chan *NotificationResult, 1),
	}

	select {
	case s.queue <- req:
		log.Printf("Queued notification: %s for build %d", config.Type, data.BuildID)
	default:
		log.Printf("Warning: Notification queue full, dropping notification")
	}
}

// NotifyBuild sends notifications for a build event
func (s *Service) NotifyBuild(ctx context.Context, projectID, userID int, data *NotificationData) []*NotificationResult {
	var results []*NotificationResult

	// Get project-specific notifications
	projectConfigs, err := s.repo.GetConfigsByProject(ctx, projectID)
	if err != nil {
		log.Printf("Failed to get project notifications: %v", err)
	}

	// Get user-global notifications
	userConfigs, err := s.repo.GetConfigsByUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user notifications: %v", err)
	}

	// Combine and deduplicate configs
	configs := append(projectConfigs, userConfigs...)

	// Send notifications
	for _, config := range configs {
		result := s.SendNotification(ctx, config, data)
		results = append(results, result)
	}

	return results
}

// worker processes notifications from the queue
func (s *Service) worker() {
	for req := range s.queue {
		ctx := context.Background()
		result := s.SendNotification(ctx, req.Config, req.Data)
		req.Result <- result
		close(req.Result)
	}
}

// matchesEvent checks if an event matches the configured events
func (s *Service) matchesEvent(configEvents []EventType, event EventType) bool {
	if len(configEvents) == 0 {
		return true // No filter means all events
	}

	for _, e := range configEvents {
		if e == event {
			return true
		}
	}
	return false
}

// evaluateFilter evaluates a filter expression
func (s *Service) evaluateFilter(filter string, data *NotificationData) (bool, error) {
	// Simple filter evaluation
	// TODO: Implement proper expression evaluation
	// For now, support basic conditions like:
	// - status == "success"
	// - branch == "main"
	// - status != "success"

	switch filter {
	case `status == "success"`:
		return data.Status == "success", nil
	case `status == "failure"`:
		return data.Status == "failure", nil
	case `branch == "main"`:
		return data.Branch == "main", nil
	case `branch != "main"`:
		return data.Branch != "main", nil
	default:
		return true, nil // Unknown filters pass by default
	}
}

// RenderTemplate renders a notification template with data
func (s *Service) RenderTemplate(templateName string, data *NotificationData) (string, error) {
	tmpl := s.templates.Lookup(templateName)
	if tmpl == nil {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// RenderCustomTemplate renders a custom template string
func (s *Service) RenderCustomTemplate(templateStr string, data *NotificationData) (string, error) {
	tmpl, err := template.New("custom").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("template parse failed: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// logNotification logs a sent notification
func (s *Service) logNotification(configID, buildID int, event EventType, result *NotificationResult) {
	ctx := context.Background()
	log := &NotificationLog{
		ConfigID: configID,
		BuildID:  buildID,
		Event:    event,
		Success:  result.Success,
		Error:    result.Error,
		SentAt:   result.SentAt,
		Duration: result.Duration,
	}

	if err := s.repo.LogNotification(ctx, log); err != nil {
		log := fmt.Sprintf("Failed to log notification: %v", err)
		fmt.Println(log)
	}
}

// CreateConfig creates a new notification configuration
func (s *Service) CreateConfig(ctx context.Context, config *NotificationConfig) error {
	return s.repo.CreateConfig(ctx, config)
}

// GetConfig retrieves a notification configuration
func (s *Service) GetConfig(ctx context.Context, id int) (*NotificationConfig, error) {
	return s.repo.GetConfig(ctx, id)
}

// GetConfigsByUser retrieves all notification configs for a user
func (s *Service) GetConfigsByUser(ctx context.Context, userID int) ([]*NotificationConfig, error) {
	return s.repo.GetConfigsByUser(ctx, userID)
}

// GetConfigsByProject retrieves all notification configs for a project
func (s *Service) GetConfigsByProject(ctx context.Context, projectID int) ([]*NotificationConfig, error) {
	return s.repo.GetConfigsByProject(ctx, projectID)
}

// UpdateConfig updates a notification configuration
func (s *Service) UpdateConfig(ctx context.Context, config *NotificationConfig) error {
	return s.repo.UpdateConfig(ctx, config)
}

// DeleteConfig deletes a notification configuration
func (s *Service) DeleteConfig(ctx context.Context, id int) error {
	return s.repo.DeleteConfig(ctx, id)
}

// GetNotificationLogs retrieves notification logs for a build
func (s *Service) GetNotificationLogs(ctx context.Context, buildID int) ([]*NotificationLog, error) {
	return s.repo.GetNotificationLogs(ctx, buildID)
}

// TestNotification sends a test notification
func (s *Service) TestNotification(ctx context.Context, config *NotificationConfig) *NotificationResult {
	testData := &NotificationData{
		Event:       EventBuildSuccess,
		ProjectName: "Test Project",
		BuildNumber: 1,
		Branch:      "main",
		Commit:      "abc123def456",
		CommitShort: "abc123d",
		Author:      "Test User",
		Status:      "success",
		Duration:    "2m 30s",
		StartedAt:   time.Now().Add(-2*time.Minute - 30*time.Second),
		FinishedAt:  time.Now(),
		URL:         "https://casci.example.com/builds/1",
		LogsURL:     "https://casci.example.com/builds/1/log",
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	return s.SendNotification(ctx, config, testData)
}

// Close closes the notification service
func (s *Service) Close() {
	close(s.queue)
}

// BuildNotificationData creates notification data from build information
func BuildNotificationData(event EventType, buildInfo map[string]interface{}) *NotificationData {
	data := &NotificationData{
		Event:    event,
		Metadata: make(map[string]interface{}),
	}

	// Extract fields from buildInfo
	if v, ok := buildInfo["project_name"].(string); ok {
		data.ProjectName = v
	}
	if v, ok := buildInfo["project_id"].(int); ok {
		data.ProjectID = v
	}
	if v, ok := buildInfo["build_id"].(int); ok {
		data.BuildID = v
	}
	if v, ok := buildInfo["build_number"].(int); ok {
		data.BuildNumber = v
	}
	if v, ok := buildInfo["branch"].(string); ok {
		data.Branch = v
	}
	if v, ok := buildInfo["commit"].(string); ok {
		data.Commit = v
		if len(v) > 7 {
			data.CommitShort = v[:7]
		} else {
			data.CommitShort = v
		}
	}
	if v, ok := buildInfo["author"].(string); ok {
		data.Author = v
	}
	if v, ok := buildInfo["status"].(string); ok {
		data.Status = v
	}
	if v, ok := buildInfo["duration"].(string); ok {
		data.Duration = v
	}
	if v, ok := buildInfo["started_at"].(time.Time); ok {
		data.StartedAt = v
	}
	if v, ok := buildInfo["finished_at"].(time.Time); ok {
		data.FinishedAt = v
	}
	if v, ok := buildInfo["url"].(string); ok {
		data.URL = v
	}
	if v, ok := buildInfo["logs_url"].(string); ok {
		data.LogsURL = v
	}
	if v, ok := buildInfo["error"].(string); ok {
		data.Error = v
	}

	return data
}

// Helper function to parse JSON config
func ParseConfigJSON(configJSON map[string]string, target interface{}) error {
	data, err := json.Marshal(configJSON)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
