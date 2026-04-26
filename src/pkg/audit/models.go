package audit

import "time"

type AuditEvent struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID int64     `json:"resource_id,omitempty"`
	Details    string    `json:"details,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

type AuditFilter struct {
	UserID     *int64
	Action     string
	Resource   string
	StartTime  *time.Time
	EndTime    *time.Time
	Success    *bool
	Limit      int
	Offset     int
}

const (
	ActionUserRegister     = "user.register"
	ActionUserLogin        = "user.login"
	ActionUserLogout       = "user.logout"
	ActionUserUpdate       = "user.update"
	ActionUserDelete       = "user.delete"
	ActionUserTokenRegen   = "user.token.regenerate"
	
	ActionProjectCreate    = "project.create"
	ActionProjectRead      = "project.read"
	ActionProjectUpdate    = "project.update"
	ActionProjectDelete    = "project.delete"
	
	ActionBuildTrigger     = "build.trigger"
	ActionBuildCancel      = "build.cancel"
	ActionBuildRestart     = "build.restart"
	ActionBuildViewLog     = "build.view_log"
	
	ActionCredentialCreate = "credential.create"
	ActionCredentialRead   = "credential.read"
	ActionCredentialUpdate = "credential.update"
	ActionCredentialDelete = "credential.delete"
	
	ActionNodeRegister     = "node.register"
	ActionNodeUpdate       = "node.update"
	ActionNodeDelete       = "node.delete"
	ActionNodeDrain        = "node.drain"
	
	ActionSecurityScan     = "security.scan"
	ActionSecurityReport   = "security.report"
	
	ActionSettingsUpdate   = "settings.update"
	ActionSettingsRead     = "settings.read"
	
	ResourceUser           = "user"
	ResourceProject        = "project"
	ResourceBuild          = "build"
	ResourceCredential     = "credential"
	ResourceNode           = "node"
	ResourceSettings       = "settings"
	ResourceSecurity       = "security"
)
