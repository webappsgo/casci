package notifications

import (
	"time"
)

// NotificationType represents the type of notification channel
type NotificationType string

const (
	// Chat/Messaging
	NotificationTypeSlack     NotificationType = "slack"
	NotificationTypeDiscord   NotificationType = "discord"
	NotificationTypeTeams     NotificationType = "teams"
	NotificationTypeTelegram  NotificationType = "telegram"
	NotificationTypeMatrix    NotificationType = "matrix"
	NotificationTypeIRC       NotificationType = "irc"
	NotificationTypeMattermost NotificationType = "mattermost"

	// Email
	NotificationTypeEmail NotificationType = "email"

	// Issue Tracking
	NotificationTypeJira   NotificationType = "jira"
	NotificationTypeGithub NotificationType = "github_issue"
	NotificationTypeGitlab NotificationType = "gitlab_issue"
	NotificationTypeLinear NotificationType = "linear"
	NotificationTypeAsana  NotificationType = "asana"

	// Git Providers
	NotificationTypeGithubStatus   NotificationType = "github_status"
	NotificationTypeGitlabStatus   NotificationType = "gitlab_status"
	NotificationTypeBitbucketStatus NotificationType = "bitbucket_status"
	NotificationTypeGiteaStatus    NotificationType = "gitea_status"

	// Incident Management
	NotificationTypePagerDuty NotificationType = "pagerduty"
	NotificationTypeOpsGenie  NotificationType = "opsgenie"
	NotificationTypeVictorOps NotificationType = "victorops"

	// Custom
	NotificationTypeWebhook NotificationType = "webhook"
	NotificationTypeScript  NotificationType = "script"
)

// EventType represents the type of build event
type EventType string

const (
	EventBuildStarted   EventType = "build.started"
	EventBuildSuccess   EventType = "build.success"
	EventBuildFailure   EventType = "build.failure"
	EventBuildCancelled EventType = "build.cancelled"
	EventDeployStarted  EventType = "deploy.started"
	EventDeploySuccess  EventType = "deploy.success"
	EventDeployFailure  EventType = "deploy.failure"
	EventSecurityAlert  EventType = "security.alert"
)

// NotificationConfig represents a notification channel configuration
type NotificationConfig struct {
	ID        int              `json:"id"`
	UserID    int              `json:"user_id"`
	ProjectID *int             `json:"project_id,omitempty"` // nil = global
	Name      string           `json:"name"`
	Type      NotificationType `json:"type"`
	Enabled   bool             `json:"enabled"`
	Config    map[string]string `json:"config"` // Type-specific configuration
	Events    []EventType      `json:"events"`  // Which events to notify on
	Filter    string           `json:"filter,omitempty"` // Conditional filter
	Template  string           `json:"template,omitempty"` // Custom message template
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NotificationData represents data passed to notification templates
type NotificationData struct {
	Event       EventType          `json:"event"`
	ProjectName string             `json:"project_name"`
	ProjectID   int                `json:"project_id"`
	BuildID     int                `json:"build_id"`
	BuildNumber int                `json:"build_number"`
	Branch      string             `json:"branch"`
	Commit      string             `json:"commit"`
	CommitShort string             `json:"commit_short"`
	Author      string             `json:"author"`
	Status      string             `json:"status"`
	Duration    string             `json:"duration"`
	StartedAt   time.Time          `json:"started_at"`
	FinishedAt  time.Time          `json:"finished_at"`
	URL         string             `json:"url"`
	LogsURL     string             `json:"logs_url"`
	Artifacts   []string           `json:"artifacts,omitempty"`
	Error       string             `json:"error,omitempty"`
	Changes     []CommitChange     `json:"changes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CommitChange represents a single commit in the change list
type CommitChange struct {
	SHA     string    `json:"sha"`
	ShortSHA string   `json:"short_sha"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Time    time.Time `json:"time"`
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	ConfigID  int       `json:"config_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	SentAt    time.Time `json:"sent_at"`
	Duration  time.Duration `json:"duration"`
}

// SlackConfig represents Slack-specific configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
	IconEmoji  string `json:"icon_emoji,omitempty"`
}

// EmailConfig represents email-specific configuration
type EmailConfig struct {
	SMTPHost     string   `json:"smtp_host"`
	SMTPPort     int      `json:"smtp_port"`
	SMTPUser     string   `json:"smtp_user"`
	SMTPPassword string   `json:"smtp_password"`
	From         string   `json:"from"`
	To           []string `json:"to"`
	Subject      string   `json:"subject,omitempty"`
	UseHTML      bool     `json:"use_html"`
}

// WebhookConfig represents webhook-specific configuration
type WebhookConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"` // Default: POST
	Headers     map[string]string `json:"headers,omitempty"`
	ContentType string            `json:"content_type"` // Default: application/json
}

// GitHubStatusConfig represents GitHub status API configuration
type GitHubStatusConfig struct {
	Token      string `json:"token"`
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	Context    string `json:"context"`    // Status context name
	TargetURL  string `json:"target_url"` // Link back to build
}

// NotificationLog represents a log entry for sent notifications
type NotificationLog struct {
	ID        int              `json:"id"`
	ConfigID  int              `json:"config_id"`
	BuildID   int              `json:"build_id"`
	Event     EventType        `json:"event"`
	Success   bool             `json:"success"`
	Error     string           `json:"error,omitempty"`
	SentAt    time.Time        `json:"sent_at"`
	Duration  time.Duration    `json:"duration"`
}

// Template constants for default templates
const (
	// Default Slack template
	DefaultSlackTemplate = `{
		"text": "{{if eq .Status "success"}}✅{{else if eq .Status "failure"}}❌{{else}}🔄{{end}} Build {{.Status}}",
		"attachments": [{
			"color": "{{if eq .Status "success"}}good{{else if eq .Status "failure"}}danger{{else}}warning{{end}}",
			"fields": [
				{"title": "Project", "value": "{{.ProjectName}}", "short": true},
				{"title": "Branch", "value": "{{.Branch}}", "short": true},
				{"title": "Commit", "value": "{{.CommitShort}} by {{.Author}}", "short": false},
				{"title": "Duration", "value": "{{.Duration}}", "short": true},
				{"title": "Build", "value": "<{{.URL}}|#{{.BuildNumber}}>", "short": true}
			]
		}]
	}`

	// Default email subject template
	DefaultEmailSubject = `[CASCI] {{.ProjectName}} - Build {{if eq .Status "success"}}✅ Success{{else if eq .Status "failure"}}❌ Failed{{else}}🔄 {{.Status}}{{end}}`

	// Default email body template (plain text)
	DefaultEmailBody = `Build {{.Status}}

Project: {{.ProjectName}}
Branch: {{.Branch}}
Commit: {{.Commit}} by {{.Author}}
Duration: {{.Duration}}

View Build: {{.URL}}
View Logs: {{.LogsURL}}

{{if .Error}}
Error: {{.Error}}
{{end}}

{{if .Changes}}
Recent Changes:
{{range .Changes}}
- {{.ShortSHA}} {{.Message}} ({{.Author}})
{{end}}
{{end}}
`

	// Default webhook template (JSON)
	DefaultWebhookTemplate = `{
		"event": "{{.Event}}",
		"project": "{{.ProjectName}}",
		"build_number": {{.BuildNumber}},
		"status": "{{.Status}}",
		"branch": "{{.Branch}}",
		"commit": "{{.Commit}}",
		"author": "{{.Author}}",
		"duration": "{{.Duration}}",
		"url": "{{.URL}}",
		"timestamp": "{{.FinishedAt.Format "2006-01-02T15:04:05Z07:00"}}"
	}`
)
