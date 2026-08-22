package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/webappsgo/casci/src/pkg/builds"
	"github.com/webappsgo/casci/src/pkg/projects"
)

// Handler handles webhook requests from various Git providers
type Handler struct {
	projectService *projects.Service
	buildService   *builds.Service
	// Project ID -> Secret
	secrets map[string]string
	// Server-wide fallback secret, used when a project has no secret of its own
	defaultSecret string
}

// NewHandler creates a new webhook handler. The shared secret comes from
// server.yml (server.security.webhook_secret); an empty secret rejects every
// inbound delivery rather than skipping verification.
func NewHandler(projectService *projects.Service, buildService *builds.Service, defaultSecret string) *Handler {
	return &Handler{
		projectService: projectService,
		buildService:   buildService,
		secrets:        make(map[string]string),
		defaultSecret:  defaultSecret,
	}
}

// webhookSecret returns the configured shared secret, or an empty string when
// none is configured. Callers MUST treat an empty secret as "reject", never as
// "skip verification".
func (h *Handler) webhookSecret() string {
	return h.defaultSecret
}

// ServeHTTP handles incoming webhook requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Detect provider
	provider := h.detectProvider(r)
	log.Printf("Received webhook from %s", provider)

	switch provider {
	case "github":
		h.handleGitHub(w, r)
	case "gitlab":
		h.handleGitLab(w, r)
	case "bitbucket":
		h.handleBitbucket(w, r)
	case "gitea":
		h.handleGitea(w, r)
	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
	}
}

// detectProvider detects the Git provider from headers
func (h *Handler) detectProvider(r *http.Request) string {
	if r.Header.Get("X-GitHub-Event") != "" {
		return "github"
	}
	if r.Header.Get("X-GitLab-Event") != "" {
		return "gitlab"
	}
	if r.Header.Get("X-Event-Key") != "" {
		return "bitbucket"
	}
	if r.Header.Get("X-Gitea-Event") != "" {
		return "gitea"
	}
	return "unknown"
}

// GitHubPushPayload represents a GitHub push webhook payload
type GitHubPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"head_commit"`
}

// GitHubPullRequestPayload represents a GitHub PR webhook payload
type GitHubPullRequestPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Number int `json:"number"`
		Head   struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
}

// handleGitHub handles GitHub webhooks
func (h *Handler) handleGitHub(w http.ResponseWriter, r *http.Request) {
	// Signature verification is mandatory. Treating a missing header as
	// "no secret configured" would let any caller skip verification entirely
	// by simply omitting the header.
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		log.Printf("Rejected GitHub webhook with no X-Hub-Signature-256 header")
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	if !h.verifyGitHubSignature(signature, body) {
		log.Printf("Invalid GitHub signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	log.Printf("GitHub event: %s", event)

	switch event {
	case "push":
		h.handleGitHubPush(w, r)
	case "pull_request":
		h.handleGitHubPullRequest(w, r)
	case "ping":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
	default:
		log.Printf("Unsupported GitHub event: %s", event)
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) handleGitHubPush(w http.ResponseWriter, r *http.Request) {
	var payload GitHubPushPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Extract branch from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Find project by repository URL
	ctx := context.Background()
	project, err := h.findProjectByURL(ctx, payload.Repository.CloneURL)
	if err != nil {
		log.Printf("Project not found for URL %s: %v", payload.Repository.CloneURL, err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Trigger build
	build := &builds.Build{
		ProjectID:     project.ID,
		Branch:        branch,
		CommitSHA:     payload.HeadCommit.ID,
		CommitMessage: payload.HeadCommit.Message,
		Status:        "queued",
		Trigger:       "push",
	}

	if err := h.buildService.Create(ctx, build); err != nil {
		log.Printf("Failed to create build: %v", err)
		http.Error(w, "Failed to create build", http.StatusInternalServerError)
		return
	}

	log.Printf("Triggered build #%d for project %s (commit: %s)", build.BuildNumber, project.Name, build.CommitSHA)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"build_id":     build.ID,
		"build_number": build.BuildNumber,
	})
}

func (h *Handler) handleGitHubPullRequest(w http.ResponseWriter, r *http.Request) {
	var payload GitHubPullRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Only trigger on opened and synchronized (new commits)
	if payload.Action != "opened" && payload.Action != "synchronize" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Find project
	ctx := context.Background()
	project, err := h.findProjectByURL(ctx, payload.Repository.CloneURL)
	if err != nil {
		log.Printf("Project not found for URL %s: %v", payload.Repository.CloneURL, err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Trigger PR build
	build := &builds.Build{
		ProjectID:     project.ID,
		Branch:        payload.PullRequest.Head.Ref,
		CommitSHA:     payload.PullRequest.Head.SHA,
		CommitMessage: fmt.Sprintf("PR #%d", payload.PullRequest.Number),
		Status:        "queued",
		Trigger:       "pull_request",
	}

	if err := h.buildService.Create(ctx, build); err != nil {
		log.Printf("Failed to create build: %v", err)
		http.Error(w, "Failed to create build", http.StatusInternalServerError)
		return
	}

	log.Printf("Triggered PR build #%d for project %s (PR #%d)", build.BuildNumber, project.Name, payload.PullRequest.Number)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"build_id":     build.ID,
		"build_number": build.BuildNumber,
	})
}

func (h *Handler) verifyGitHubSignature(signature string, body []byte) bool {
	// Fail closed: with no secret configured there is nothing to verify
	// against, so no signature can be considered valid.
	secret := []byte(h.webhookSecret())
	if len(secret) == 0 {
		log.Printf("Webhook signature verification failed: no webhook secret configured")
		return false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC)

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GitLabPushPayload represents a GitLab push webhook payload
type GitLabPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		GitHTTPURL string `json:"git_http_url"`
	} `json:"repository"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

// handleGitLab handles GitLab webhooks
func (h *Handler) handleGitLab(w http.ResponseWriter, r *http.Request) {
	// Fail closed: a GitLab webhook is only trusted when its token matches the
	// configured secret. Compared in constant time to avoid a timing oracle.
	token := r.Header.Get("X-GitLab-Token")
	secret := h.webhookSecret()
	if secret == "" || subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
		log.Printf("Rejected GitLab webhook with missing or invalid X-GitLab-Token")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-GitLab-Event")
	log.Printf("GitLab event: %s", event)

	switch event {
	case "Push Hook":
		h.handleGitLabPush(w, r)
	case "Merge Request Hook":
		h.handleGitLabMergeRequest(w, r)
	default:
		log.Printf("Unsupported GitLab event: %s", event)
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) handleGitLabPush(w http.ResponseWriter, r *http.Request) {
	var payload GitLabPushPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Extract branch
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Find project
	ctx := context.Background()
	project, err := h.findProjectByURL(ctx, payload.Repository.GitHTTPURL)
	if err != nil {
		log.Printf("Project not found for URL %s: %v", payload.Repository.GitHTTPURL, err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Get last commit
	if len(payload.Commits) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	lastCommit := payload.Commits[len(payload.Commits)-1]

	// Trigger build
	build := &builds.Build{
		ProjectID:     project.ID,
		Branch:        branch,
		CommitSHA:     lastCommit.ID,
		CommitMessage: lastCommit.Message,
		Status:        "queued",
		Trigger:       "push",
	}

	if err := h.buildService.Create(ctx, build); err != nil {
		log.Printf("Failed to create build: %v", err)
		http.Error(w, "Failed to create build", http.StatusInternalServerError)
		return
	}

	log.Printf("Triggered build #%d for project %s", build.BuildNumber, project.Name)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"build_id":     build.ID,
		"build_number": build.BuildNumber,
	})
}

func (h *Handler) handleGitLabMergeRequest(w http.ResponseWriter, r *http.Request) {
	// Similar to GitHub PR handling
	// TODO: Implement GitLab merge request handling
	w.WriteHeader(http.StatusOK)
}

// handleBitbucket handles Bitbucket webhooks
func (h *Handler) handleBitbucket(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-Event-Key")
	log.Printf("Bitbucket event: %s", event)

	switch event {
	case "repo:push":
		// TODO: Implement Bitbucket push handling
		w.WriteHeader(http.StatusOK)
	case "pullrequest:created", "pullrequest:updated":
		// TODO: Implement Bitbucket PR handling
		w.WriteHeader(http.StatusOK)
	default:
		log.Printf("Unsupported Bitbucket event: %s", event)
		w.WriteHeader(http.StatusOK)
	}
}

// handleGitea handles Gitea webhooks
func (h *Handler) handleGitea(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-Gitea-Event")
	log.Printf("Gitea event: %s", event)

	switch event {
	case "push":
		// Gitea uses GitHub-compatible payloads
		h.handleGitHubPush(w, r)
	case "pull_request":
		h.handleGitHubPullRequest(w, r)
	default:
		log.Printf("Unsupported Gitea event: %s", event)
		w.WriteHeader(http.StatusOK)
	}
}

// ErrProjectNotFound is returned when no project matches a webhook's repository
// URL.
var ErrProjectNotFound = errors.New("no project matches repository URL")

// canonicalRepoURL reduces a repository URL to "host/path" so the clone-URL
// spellings that refer to one repository all compare equal. A project may be
// registered as https://github.com/user/repo while the webhook delivers
// git@github.com:user/repo.git, so matching on the raw string would miss.
// Comparison is case-insensitive because hosts are case-insensitive and the
// forges this talks to treat owner and repository names that way too.
func canonicalRepoURL(repoURL string) string {
	s := strings.TrimSpace(repoURL)

	// scp-style SSH ("git@host:owner/repo") is not a URL, so convert the first
	// colon after the host into a path separator before any parsing.
	if !strings.Contains(s, "://") {
		if at := strings.Index(s, "@"); at >= 0 {
			s = s[at+1:]
		}
		s = strings.Replace(s, ":", "/", 1)
	} else {
		s = s[strings.Index(s, "://")+3:]
		if at := strings.Index(s, "@"); at >= 0 {
			s = s[at+1:]
		}
	}

	// Drop any port, query, or fragment so they cannot split a match.
	if i := strings.IndexAny(s, "?#"); i >= 0 {
		s = s[:i]
	}
	if host, rest, ok := strings.Cut(s, "/"); ok {
		if h, _, found := strings.Cut(host, ":"); found {
			host = h
		}
		s = host + "/" + rest
	}

	s = strings.TrimSuffix(s, "/")
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimSuffix(s, "/")

	return strings.ToLower(s)
}

// findProjectByURL finds the project whose repository URL refers to the same
// repository as repoURL.
func (h *Handler) findProjectByURL(ctx context.Context, repoURL string) (*projects.Project, error) {
	want := canonicalRepoURL(repoURL)
	if want == "" {
		return nil, ErrProjectNotFound
	}

	// There is no repository-URL index to query: the stored spelling and the
	// delivered spelling differ often enough that an equality lookup would miss
	// most matches, so candidates are canonicalised in Go instead.
	all, err := h.projectService.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, p := range all {
		if canonicalRepoURL(p.RepositoryURL) == want {
			return p, nil
		}
	}

	return nil, ErrProjectNotFound
}
