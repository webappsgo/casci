package notifications

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"time"
)

// SlackSender sends notifications to Slack
type SlackSender struct{}

func NewSlackSender() *SlackSender {
	return &SlackSender{}
}

func (s *SlackSender) Type() NotificationType {
	return NotificationTypeSlack
}

func (s *SlackSender) Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error {
	webhookURL, ok := config.Config["webhook_url"]
	if !ok {
		return fmt.Errorf("webhook_url not configured")
	}

	// Build Slack message
	message := s.buildMessage(config, data)

	// Send to Slack webhook
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *SlackSender) buildMessage(config *NotificationConfig, data *NotificationData) map[string]interface{} {
	emoji := "🔄"
	color := "warning"

	switch data.Status {
	case "success":
		emoji = "✅"
		color = "good"
	case "failure":
		emoji = "❌"
		color = "danger"
	case "cancelled":
		emoji = "⛔"
		color = "#808080"
	}

	message := map[string]interface{}{
		"text": fmt.Sprintf("%s Build %s", emoji, data.Status),
	}

	// Add channel if specified
	if channel, ok := config.Config["channel"]; ok && channel != "" {
		message["channel"] = channel
	}

	// Add username if specified
	if username, ok := config.Config["username"]; ok && username != "" {
		message["username"] = username
	} else {
		message["username"] = "CASCI"
	}

	// Add icon if specified
	if icon, ok := config.Config["icon_emoji"]; ok && icon != "" {
		message["icon_emoji"] = icon
	}

	// Build attachments
	fields := []map[string]interface{}{
		{"title": "Project", "value": data.ProjectName, "short": true},
		{"title": "Branch", "value": data.Branch, "short": true},
		{"title": "Commit", "value": fmt.Sprintf("%s by %s", data.CommitShort, data.Author), "short": false},
		{"title": "Duration", "value": data.Duration, "short": true},
		{"title": "Build", "value": fmt.Sprintf("<%s|#%d>", data.URL, data.BuildNumber), "short": true},
	}

	if data.Error != "" {
		fields = append(fields, map[string]interface{}{
			"title": "Error",
			"value": data.Error,
			"short": false,
		})
	}

	message["attachments"] = []map[string]interface{}{
		{
			"color":  color,
			"fields": fields,
		},
	}

	return message
}

// DiscordSender sends notifications to Discord
type DiscordSender struct{}

func NewDiscordSender() *DiscordSender {
	return &DiscordSender{}
}

func (d *DiscordSender) Type() NotificationType {
	return NotificationTypeDiscord
}

func (d *DiscordSender) Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error {
	webhookURL, ok := config.Config["webhook_url"]
	if !ok {
		return fmt.Errorf("webhook_url not configured")
	}

	// Build Discord embed
	embed := d.buildEmbed(data)

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	if username, ok := config.Config["username"]; ok && username != "" {
		payload["username"] = username
	}

	// Send to Discord webhook
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (d *DiscordSender) buildEmbed(data *NotificationData) map[string]interface{} {
	color := 0xFFA500 // Orange for in-progress

	switch data.Status {
	case "success":
		color = 0x00FF00 // Green
	case "failure":
		color = 0xFF0000 // Red
	case "cancelled":
		color = 0x808080 // Gray
	}

	embed := map[string]interface{}{
		"title":       fmt.Sprintf("Build #%d - %s", data.BuildNumber, data.Status),
		"url":         data.URL,
		"color":       color,
		"timestamp":   data.FinishedAt.Format(time.RFC3339),
		"description": fmt.Sprintf("**%s** on branch **%s**", data.ProjectName, data.Branch),
		"fields": []map[string]interface{}{
			{"name": "Commit", "value": fmt.Sprintf("%s by %s", data.CommitShort, data.Author), "inline": true},
			{"name": "Duration", "value": data.Duration, "inline": true},
		},
	}

	if data.Error != "" {
		fields := embed["fields"].([]map[string]interface{})
		fields = append(fields, map[string]interface{}{
			"name":   "Error",
			"value":  data.Error,
			"inline": false,
		})
		embed["fields"] = fields
	}

	return embed
}

// EmailSender sends notifications via SMTP
type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (e *EmailSender) Type() NotificationType {
	return NotificationTypeEmail
}

func (e *EmailSender) Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error {
	// Parse email configuration
	var emailConfig EmailConfig
	if err := ParseConfigJSON(config.Config, &emailConfig); err != nil {
		return fmt.Errorf("invalid email configuration: %w", err)
	}

	// Build subject and body
	subject := e.buildSubject(data)
	if emailConfig.Subject != "" {
		subject = emailConfig.Subject
	}

	body := e.buildBody(data)

	// Build email
	message := fmt.Sprintf("From: %s\r\n", emailConfig.From)
	message += fmt.Sprintf("To: %s\r\n", emailConfig.To[0])
	message += fmt.Sprintf("Subject: %s\r\n", subject)

	if emailConfig.UseHTML {
		message += "MIME-Version: 1.0\r\n"
		message += "Content-Type: text/html; charset=utf-8\r\n"
	}

	message += "\r\n" + body

	// Send email
	auth := smtp.PlainAuth("", emailConfig.SMTPUser, emailConfig.SMTPPassword, emailConfig.SMTPHost)
	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)

	err := smtp.SendMail(addr, auth, emailConfig.From, emailConfig.To, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (e *EmailSender) buildSubject(data *NotificationData) string {
	emoji := "🔄"
	switch data.Status {
	case "success":
		emoji = "✅"
	case "failure":
		emoji = "❌"
	}

	return fmt.Sprintf("[CASCI] %s - Build %s %s", data.ProjectName, emoji, data.Status)
}

func (e *EmailSender) buildBody(data *NotificationData) string {
	body := fmt.Sprintf("Build %s\n\n", data.Status)
	body += fmt.Sprintf("Project: %s\n", data.ProjectName)
	body += fmt.Sprintf("Branch: %s\n", data.Branch)
	body += fmt.Sprintf("Commit: %s by %s\n", data.Commit, data.Author)
	body += fmt.Sprintf("Duration: %s\n\n", data.Duration)
	body += fmt.Sprintf("View Build: %s\n", data.URL)
	body += fmt.Sprintf("View Logs: %s\n", data.LogsURL)

	if data.Error != "" {
		body += fmt.Sprintf("\nError: %s\n", data.Error)
	}

	if len(data.Changes) > 0 {
		body += "\nRecent Changes:\n"
		for _, change := range data.Changes {
			body += fmt.Sprintf("- %s %s (%s)\n", change.ShortSHA, change.Message, change.Author)
		}
	}

	return body
}

// WebhookSender sends generic webhook notifications
type WebhookSender struct{}

func NewWebhookSender() *WebhookSender {
	return &WebhookSender{}
}

func (w *WebhookSender) Type() NotificationType {
	return NotificationTypeWebhook
}

func (w *WebhookSender) Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error {
	url, ok := config.Config["url"]
	if !ok {
		return fmt.Errorf("url not configured")
	}

	method := "POST"
	if m, ok := config.Config["method"]; ok && m != "" {
		method = m
	}

	contentType := "application/json"
	if ct, ok := config.Config["content_type"]; ok && ct != "" {
		contentType = ct
	}

	// Build payload
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	// Add custom headers
	for key, value := range config.Config {
		if key != "url" && key != "method" && key != "content_type" {
			req.Header.Set(key, value)
		}
	}

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GitHubStatusSender sends build status to GitHub
type GitHubStatusSender struct{}

func NewGitHubStatusSender() *GitHubStatusSender {
	return &GitHubStatusSender{}
}

func (g *GitHubStatusSender) Type() NotificationType {
	return NotificationTypeGithubStatus
}

func (g *GitHubStatusSender) Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error {
	token, ok := config.Config["token"]
	if !ok {
		return fmt.Errorf("token not configured")
	}

	owner, ok := config.Config["owner"]
	if !ok {
		return fmt.Errorf("owner not configured")
	}

	repo, ok := config.Config["repo"]
	if !ok {
		return fmt.Errorf("repo not configured")
	}

	// Map CASCI status to GitHub status
	state := "pending"
	description := "Build is running"

	switch data.Status {
	case "success":
		state = "success"
		description = fmt.Sprintf("Build passed in %s", data.Duration)
	case "failure":
		state = "failure"
		description = "Build failed"
		if data.Error != "" {
			description = fmt.Sprintf("Build failed: %s", data.Error)
		}
	case "cancelled":
		state = "error"
		description = "Build was cancelled"
	}

	context_name := "CASCI"
	if ctx_name, ok := config.Config["context"]; ok && ctx_name != "" {
		context_name = ctx_name
	}

	// Build GitHub status payload
	payload := map[string]interface{}{
		"state":       state,
		"target_url":  data.URL,
		"description": description,
		"context":     context_name,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send to GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/statuses/%s", owner, repo, data.Commit)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GitLabStatusSender sends build status to GitLab
type GitLabStatusSender struct{}

func NewGitLabStatusSender() *GitLabStatusSender {
	return &GitLabStatusSender{}
}

func (g *GitLabStatusSender) Type() NotificationType {
	return NotificationTypeGitlabStatus
}

func (g *GitLabStatusSender) Send(ctx context.Context, config *NotificationConfig, data *NotificationData) error {
	token, ok := config.Config["token"]
	if !ok {
		return fmt.Errorf("token not configured")
	}

	projectID, ok := config.Config["project_id"]
	if !ok {
		return fmt.Errorf("project_id not configured")
	}

	// Map CASCI status to GitLab status
	state := "running"

	switch data.Status {
	case "success":
		state = "success"
	case "failure":
		state = "failed"
	case "cancelled":
		state = "canceled"
	}

	// Build GitLab status payload
	payload := map[string]interface{}{
		"state":      state,
		"target_url": data.URL,
		"name":       "CASCI",
	}

	if data.Error != "" {
		payload["description"] = data.Error
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send to GitLab API
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/statuses/%s", projectID, data.Commit)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
