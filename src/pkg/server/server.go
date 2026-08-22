package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/webappsgo/casci/src/internal/config"
	"github.com/webappsgo/casci/src/pkg/audit"
	"github.com/webappsgo/casci/src/pkg/builds"
	"github.com/webappsgo/casci/src/pkg/compliance"
	"github.com/webappsgo/casci/src/pkg/credentials"
	"github.com/webappsgo/casci/src/pkg/database"
	"github.com/webappsgo/casci/src/pkg/executor"
	"github.com/webappsgo/casci/src/pkg/graphql"
	"github.com/webappsgo/casci/src/pkg/metrics"
	"github.com/webappsgo/casci/src/pkg/nodes"
	"github.com/webappsgo/casci/src/pkg/notifications"
	"github.com/webappsgo/casci/src/pkg/projects"
	"github.com/webappsgo/casci/src/pkg/queue"
	"github.com/webappsgo/casci/src/pkg/security"
	"github.com/webappsgo/casci/src/pkg/users"
	"github.com/webappsgo/casci/src/pkg/webhooks"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/webappsgo/casci/src/pkg/server/docs"
)

// Server represents the HTTP/WebSocket server
type Server struct {
	config              *config.Config
	db                  *database.Database
	router              *http.ServeMux
	server              *http.Server
	templateRenderer    *TemplateRenderer
	userService         *users.Service
	authManager         *users.AuthManager
	userHandler         *users.Handler
	userMiddleware      *users.Middleware
	projectService      *projects.Service
	projectHandler      *projects.Handler
	buildService        *builds.Service
	buildHandler        *builds.Handler
	nodeService         *nodes.Service
	nodeHandler         *nodes.Handler
	securityService     *security.Service
	securityHandler     *security.Handler
	notificationService *notifications.Service
	notificationHandler *notifications.Handler
	credentialService   *credentials.Service
	credentialHandler   *credentials.Handler
	auditService        *audit.Service
	auditHandler        *audit.Handlers
	complianceService   *compliance.Service
	complianceHandler   *compliance.Handlers
	webhookHandler      *webhooks.Handler
	metricsCollector    *metrics.Collector
	metricsExporter     *metrics.PrometheusExporter
	metricsHandler      *metrics.Handler
	graphqlHandler      *handler.Server
	csrfManager         *CSRFManager
	executor            executor.Executor
	buildQueue          *queue.BuildQueue
	// startedAt is the process start time, reported as health-check uptime
	startedAt  time.Time
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// New creates a new server instance
func New(cfg *config.Config, db *database.Database) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	srv := &Server{
		config:     cfg,
		db:         db,
		router:     http.NewServeMux(),
		startedAt:  time.Now(),
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// Initialize template renderer (TEMPLATE.md compliant)
	// Development mode reloads templates from disk on every render
	devMode := cfg.Server.Mode == "development"
	templateRenderer, err := NewTemplateRenderer(devMode)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template renderer: %w", err)
	}
	srv.templateRenderer = templateRenderer

	// Initialize CSRF manager (TEMPLATE.md PART 13 required)
	srv.csrfManager = NewCSRFManager()

	// Initialize user management
	srv.userService = users.NewService(db)
	srv.authManager, err = users.NewAuthManager(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth manager: %w", err)
	}
	srv.userHandler = users.NewHandler(srv.userService, srv.authManager)
	srv.userMiddleware = users.NewMiddleware(srv.authManager, srv.userService)

	// Initialize project management
	srv.projectService = projects.NewService(db)
	srv.projectHandler = projects.NewHandler(srv.projectService)

	// Initialize build management
	srv.buildService = builds.NewService(db)
	srv.buildHandler = builds.NewHandler(srv.buildService, srv.projectService)

	// Initialize node management
	srv.nodeService = nodes.NewService(db)
	srv.nodeHandler = nodes.NewHandler(srv.nodeService)

	// Initialize security scanning
	securityRepo := security.NewSQLRepository(db.DB())
	securityConfig := &security.ScanConfig{
		EnableVulnScan:    true,
		EnableSAST:        true,
		EnableSecretScan:  true,
		EnableLicenseScan: true,
		EnableSBOM:        true,
		FailOnCritical:    true,
		FailOnHigh:        false,
		SBOMFormat:        "spdx",
	}
	srv.securityService = security.NewService(securityConfig, securityRepo)
	srv.securityHandler = security.NewHandler(srv.securityService)

	// Initialize notification system
	notificationRepo := notifications.NewSQLRepository(db.DB())
	srv.notificationService = notifications.NewService(notificationRepo)
	srv.notificationHandler = notifications.NewHandler(srv.notificationService)
	log.Printf("Notification service started with 10 workers")

	// Initialize metrics system
	srv.metricsCollector = metrics.NewCollector()
	srv.metricsExporter = metrics.NewPrometheusExporter(srv.metricsCollector)
	srv.metricsHandler = metrics.NewHandler(srv.metricsExporter)
	log.Printf("Metrics collector started")

	// Connect metrics to services
	srv.buildService.SetMetrics(srv.metricsCollector)

	// Initialize credential management
	credRepo := credentials.NewRepository(db.DB())
	srv.credentialService = credentials.NewService(credRepo, cfg.EncryptionKey)
	srv.credentialHandler = credentials.NewHandler(srv.credentialService)
	log.Printf("Credential management initialized")

	// Initialize audit system
	auditRepo := audit.NewSQLRepository(db.DB())
	auditConfig := &audit.Config{
		Enabled:   true,
		Retention: 90 * 24 * time.Hour, // 90 days default
	}
	srv.auditService = audit.NewService(auditRepo, auditConfig)
	srv.auditHandler = audit.NewHandlers(srv.auditService)
	go srv.auditService.StartCleanupScheduler(srv.ctx)
	log.Printf("Audit system initialized with 90-day retention")

	// Initialize compliance system
	complianceConfig := compliance.GetPresetConfig(compliance.ModeNone)
	srv.complianceService = compliance.NewService(complianceConfig)
	srv.complianceHandler = compliance.NewHandlers(srv.complianceService)
	log.Printf("Compliance system initialized")

	// Initialize webhook handler
	srv.webhookHandler = webhooks.NewHandler(srv.projectService, srv.buildService, cfg.WebhookSecret)

	// Initialize GraphQL handler
	resolver := graphql.NewResolver(
		srv.userService,
		srv.authManager,
		srv.projectService,
		srv.buildService,
		srv.nodeService,
		srv.metricsCollector,
	)
	srv.graphqlHandler = handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolver}))
	log.Printf("GraphQL handler initialized")

	// Initialize executor
	// Roots come from the PART 4 path resolver, never a hardcoded /var/lib path
	executorConfig := &executor.Config{
		Type:           executor.TypeContainer,
		WorkspaceRoot:  filepath.Join(config.DataDir(), "workspaces"),
		CacheRoot:      config.CacheDir(),
		ArtifactsRoot:  filepath.Join(config.DataDir(), "artifacts"),
		DefaultTimeout: 3600, // 1 hour default
	}

	srv.executor, err = executor.NewContainerExecutor(executorConfig, srv.securityService)
	if err != nil {
		log.Printf("Warning: Failed to initialize container executor: %v", err)
		log.Printf("Builds will fail until Docker is available")
	}

	// Initialize build queue
	srv.buildQueue = queue.NewBuildQueue(srv.buildService, srv.executor, 5)
	srv.buildQueue.Start(srv.ctx)

	log.Printf("Build queue started with 5 workers")

	// Start node health checking
	go srv.nodeService.StartHealthCheck(srv.ctx)
	log.Printf("Node health check started")

	// Setup routes
	srv.setupRoutes()

	// Every response passes through the security headers, then CORS, then CSRF
	// validation before reaching a route
	handler := srv.SecurityHeadersMiddleware(srv.CORSMiddleware(srv.CSRFMiddleware(srv.router)))

	// Create HTTP server
	srv.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           handler,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		// TLS 1.0/1.1 are not acceptable - AI.md PART 11 requires TLS 1.2+.
		// server.ssl.min_version can raise the floor further; it can never
		// lower it, since TLSVersion falls back to 1.2 for anything else.
		TLSConfig: &tls.Config{MinVersion: config.TLSVersion(cfg.Server.TLSMinVersion)},
	}

	return srv, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// API version prefix is operator-configurable (server.api_version), so no
	// route may hardcode /api/v1
	api := "/api/" + s.config.Server.APIVersion

	// Web UI Routes (TEMPLATE.md compliant - HTML endpoints)
	// "/" is the catch-all pattern: the home page on an exact match, 404 otherwise
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			s.handle404(w, r)
			return
		}
		s.handleIndex(w, r)
	})
	// Canonical health check HTML (AI.md PART 1 endpoint pattern)
	s.router.HandleFunc("/server/healthz", s.handleHealthzHTML)
	// Optional root compatibility alias, mounted to the SAME handler, never a
	// redirect. Off unless server.healthz.root.enabled is set.
	if s.config.Server.HealthzRootEnabled {
		s.router.HandleFunc("/healthz", s.handleHealthzHTML)
	}

	// Admin pages live under /server/{admin_path} and require an authenticated
	// administrator. The segment is operator-configurable, never a literal.
	adminRoot := "/server/" + s.config.Server.AdminPath
	s.router.Handle(adminRoot, s.userMiddleware.RequireAdmin(http.HandlerFunc(s.handleAdminDashboard)))
	s.router.Handle(adminRoot+"/", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.handle404)))
	s.router.Handle(adminRoot+"/config/settings", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.handleAdminSettings)))

	// Standard pages (TEMPLATE.md PART 13 required)
	// About page
	s.router.HandleFunc("/server/about", s.handleServerAbout)
	// Privacy policy
	s.router.HandleFunc("/server/privacy", s.handleServerPrivacy)
	// Contact form
	s.router.HandleFunc("/server/contact", s.handleServerContact)
	// Help/documentation
	s.router.HandleFunc("/server/help", s.handleServerHelp)

	// Well-known URIs (RFC 8615). security.txt is required on every project;
	// every other entry 404s rather than redirecting.
	s.router.HandleFunc("/.well-known/security.txt", s.handleSecurityTxt)
	s.router.HandleFunc("/.well-known/llms.txt", s.handleLLMsTxt)
	s.router.HandleFunc("/llms.txt", s.handleLLMsTxt)
	s.router.HandleFunc("/.well-known/", s.handleWellKnown)

	// Human-readable counterpart of security.txt — the Policy: target
	s.router.HandleFunc("/server/security", s.handleSecurityPolicy)

	// Static files (embedded). staticRoot is the embedded tree re-rooted at its
	// "static" directory — without that, stripping the /static/ request prefix
	// leaves a path that does not exist in the FS and every stylesheet, script
	// and image 404s.
	s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticRoot))))

	// Swagger UI (HTML) at the canonical documentation route. The exact path is
	// rewritten internally to index.html so the UI renders without a redirect.
	s.router.HandleFunc("/server/docs/swagger", func(w http.ResponseWriter, r *http.Request) {
		rewritten := r.Clone(r.Context())
		rewritten.URL.Path = "/server/docs/swagger/index.html"
		rewritten.RequestURI = rewritten.URL.RequestURI()
		httpSwagger.WrapHandler(w, rewritten)
	})
	// Swagger UI static assets
	s.router.HandleFunc("/server/docs/swagger/", httpSwagger.WrapHandler)
	// OpenAPI spec (JSON/YAML) with the documented unversioned aliases
	s.router.HandleFunc(api+"/server/swagger", s.serveOpenAPIJSON)
	s.router.HandleFunc("/api/swagger", s.serveOpenAPIJSON)
	s.router.HandleFunc(api+"/server/swagger.yaml", s.serveOpenAPIYAML)
	s.router.HandleFunc("/api/swagger.yaml", s.serveOpenAPIYAML)

	// GraphQL: GraphiQL UI on the web route, POST endpoint on the API route
	// with the documented unversioned alias
	s.router.Handle(api+"/server/graphql", s.graphqlHandler)
	s.router.Handle("/api/graphql", s.graphqlHandler)
	s.router.Handle("/server/docs/graphql", playground.Handler("CASCI GraphQL", api+"/server/graphql"))

	// Health check and metrics endpoints (public)
	// JSON health check paired with the /server/healthz HTML route, plus the
	// unversioned alias for machine probing (AI.md PART 13)
	s.router.HandleFunc(api+"/server/healthz", s.metricsHandler.ServeHealth)
	s.router.HandleFunc("/api/healthz", s.metricsHandler.ServeHealth)
	s.router.HandleFunc("/readyz", s.metricsHandler.ServeReadiness)
	s.router.HandleFunc("/livez", s.metricsHandler.ServeLiveness)
	// Metrics endpoints are INTERNAL (AI.md PART 11/21) - never public
	s.router.Handle("/metrics", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeMetrics)))
	s.router.Handle("/metrics/json", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeMetricsJSON)))
	s.router.Handle(api+"/metrics/system", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeSystemMetrics)))
	s.router.Handle(api+"/metrics/builds", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeBuildMetrics)))
	s.router.Handle(api+"/metrics/nodes", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeNodeMetrics)))
	s.router.Handle(api+"/metrics/security", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeSecurityMetrics)))
	s.router.Handle(api+"/metrics/api", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.metricsHandler.ServeAPIMetrics)))

	// Authentication endpoints (public)
	s.router.HandleFunc(api+"/server/auth/register", s.userHandler.Register)
	s.router.HandleFunc(api+"/server/auth/login", s.userHandler.Login)
	s.router.Handle(api+"/server/auth/refresh", s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.RefreshToken)))

	// User endpoints (authenticated)
	s.router.Handle(api+"/users/me", s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.GetMe)))
	s.router.Handle(api+"/users/me/token", s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.RegenerateAPIToken)))

	// User management (admin only)
	s.router.Handle(api+"/users", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.userHandler.ListUsers)))

	// Individual user operations
	s.router.HandleFunc(api+"/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.GetUser)).ServeHTTP(w, r)
		case http.MethodPut, http.MethodPatch:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.UpdateUser)).ServeHTTP(w, r)
		case http.MethodDelete:
			s.userMiddleware.RequireAdmin(http.HandlerFunc(s.userHandler.DeleteUser)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Project endpoints (authenticated)
	s.router.Handle(api+"/projects", s.userMiddleware.RequireAuth(http.HandlerFunc(s.projectHandler.List)))
	// Single registration for every /api/v1/projects/{id}/... sub-resource -
	// net/http panics if a pattern is registered twice
	s.router.Handle(api+"/projects/", s.userMiddleware.RequireAuth(http.HandlerFunc(s.handleProjectSubroutes)))

	// Single registration for every /api/v1/builds/{id}/... sub-resource
	s.router.Handle(api+"/builds/", s.userMiddleware.RequireAuth(http.HandlerFunc(s.handleBuildSubroutes)))

	// Node endpoints (authenticated)
	s.router.Handle(api+"/nodes", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.List)))
	s.router.HandleFunc(api+"/nodes/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path suffix
		if contains(r.URL.Path, "/heartbeat") {
			// Node heartbeat (nodes authenticate themselves)
			s.nodeHandler.Heartbeat(w, r)
		} else if contains(r.URL.Path, "/drain") {
			// Admin only
			s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.Drain)).ServeHTTP(w, r)
		} else {
			// CRUD operations
			switch r.Method {
			case http.MethodGet:
				s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.Get)).ServeHTTP(w, r)
			case http.MethodPut, http.MethodPatch:
				s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.Update)).ServeHTTP(w, r)
			case http.MethodDelete:
				s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.Delete)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})

	// Node registration (public - secured by token)
	s.router.HandleFunc(api+"/nodes/register", s.nodeHandler.Register)

	// Node token generation (admin only)
	s.router.Handle(api+"/nodes/token", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.GenerateToken)))

	// Security endpoints (authenticated)
	s.router.Handle(api+"/security/reports", s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.ListSecurityReports)))
	s.router.HandleFunc(api+"/security/reports/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.GetSecurityReport)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	s.router.Handle(api+"/security/statistics", s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.GetStatistics)))
	s.router.Handle(api+"/security/config", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.securityHandler.GetConfig(w, r)
		} else if r.Method == http.MethodPut {
			s.securityHandler.UpdateConfig(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Notification endpoints (authenticated)
	s.router.Handle(api+"/notifications", s.userMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.notificationHandler.ListConfigs(w, r)
		} else if r.Method == http.MethodPost {
			s.notificationHandler.CreateConfig(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	s.router.HandleFunc(api+"/notifications/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path suffix
		if contains(r.URL.Path, "/test") {
			// POST /api/v1/notifications/test - Test notification
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.TestConfig)).ServeHTTP(w, r)
		} else {
			// CRUD operations on individual configs
			switch r.Method {
			case http.MethodGet:
				s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetConfig)).ServeHTTP(w, r)
			case http.MethodPut, http.MethodPatch:
				s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.UpdateConfig)).ServeHTTP(w, r)
			case http.MethodDelete:
				s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.DeleteConfig)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})

	// Notification metadata endpoints
	s.router.Handle(api+"/notifications/types", s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetSupportedTypes)))
	s.router.Handle(api+"/notifications/events", s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetEventTypes)))

	// User credential endpoints (authenticated)
	s.router.Handle(api+"/credentials/user", s.userMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.credentialHandler.ListUserCredentials(w, r)
		} else if r.Method == http.MethodPost {
			s.credentialHandler.CreateUserCredential(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	s.router.HandleFunc(api+"/credentials/user/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.GetUserCredential)).ServeHTTP(w, r)
		case http.MethodPut, http.MethodPatch:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.UpdateUserCredential)).ServeHTTP(w, r)
		case http.MethodDelete:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.DeleteUserCredential)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.router.HandleFunc(api+"/credentials/project/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.GetProjectCredential)).ServeHTTP(w, r)
		case http.MethodPut, http.MethodPatch:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.UpdateProjectCredential)).ServeHTTP(w, r)
		case http.MethodDelete:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.DeleteProjectCredential)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// API v1 info
	s.router.HandleFunc(api+"/", s.handleAPIv1)

	// Audit endpoints (authenticated)
	s.router.Handle(api+"/audit/events", s.userMiddleware.RequireAuth(http.HandlerFunc(s.auditHandler.ListEvents)))
	s.router.Handle(api+"/audit/status", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.auditHandler.GetStatus)))
	s.router.Handle(api+"/audit/config", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			s.auditHandler.UpdateConfig(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	s.router.Handle(api+"/audit/cleanup", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.auditHandler.TriggerCleanup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Compliance endpoints (authenticated)
	s.router.Handle(api+"/compliance/config", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.complianceHandler.GetConfig(w, r)
		} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			s.complianceHandler.UpdateConfig(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	s.router.Handle(api+"/compliance/modes", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.complianceHandler.ListModes)))
	s.router.Handle(api+"/compliance/mode", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			s.complianceHandler.SetMode(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	s.router.Handle(api+"/compliance/preset", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.complianceHandler.GetPreset)))
	s.router.Handle(api+"/compliance/check", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.complianceHandler.RunCompliance(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Jenkins compatibility routes
	s.router.HandleFunc("/api/json", s.handleJenkinsAPI)
	s.router.HandleFunc("/crumbIssuer/api/json", s.handleJenkinsCrumb)

	// Webhook endpoint (public - secured by provider signatures)
	s.router.HandleFunc("/webhook", s.webhookHandler.ServeHTTP)

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.handleWebSocket)
}

// handleProjectSubroutes dispatches every /api/v1/projects/... request. All
// project sub-resources share one mux pattern, so the remaining path segments
// and the method select the handler.
func (s *Server) handleProjectSubroutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case contains(path, "/builds"):
		switch {
		case contains(path, "/builds/stats"):
			s.buildHandler.GetStats(w, r)
		case r.Method == http.MethodPost && !contains(path, "/builds/"):
			// POST /api/v1/projects/{id}/builds - Trigger build
			s.buildHandler.Trigger(w, r)
		case r.Method == http.MethodGet:
			// GET /api/v1/projects/{id}/builds - List builds
			s.buildHandler.ListByProject(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	case contains(path, "/notifications"):
		if r.Method == http.MethodGet {
			s.notificationHandler.GetProjectConfigs(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	case contains(path, "/credentials"):
		switch r.Method {
		case http.MethodGet:
			s.credentialHandler.ListProjectCredentials(w, r)
		case http.MethodPost:
			s.credentialHandler.CreateProjectCredential(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	default:
		// CRUD on the project itself
		switch r.Method {
		case http.MethodPost:
			s.projectHandler.Create(w, r)
		case http.MethodGet:
			s.projectHandler.Get(w, r)
		case http.MethodPut, http.MethodPatch:
			s.projectHandler.Update(w, r)
		case http.MethodDelete:
			s.projectHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleBuildSubroutes dispatches every /api/v1/builds/... request from the
// single mux pattern that owns that prefix.
func (s *Server) handleBuildSubroutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case contains(path, "/security/scan"):
		if r.Method == http.MethodPost {
			s.securityHandler.TriggerScan(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	case contains(path, "/security"):
		s.securityHandler.GetBuildSecurityReports(w, r)

	case contains(path, "/notifications"):
		if r.Method == http.MethodGet {
			s.notificationHandler.GetBuildLogs(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	case contains(path, "/log"):
		s.buildHandler.GetLog(w, r)

	case contains(path, "/cancel"):
		s.buildHandler.Cancel(w, r)

	case contains(path, "/restart"):
		s.buildHandler.Restart(w, r)

	case r.Method == http.MethodGet:
		s.buildHandler.Get(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s:%d", s.config.Server.Host, s.config.Server.Port)

	if s.config.Server.TLSEnabled {
		// TLSConfig pins MinVersion to TLS 1.2 (set in New)
		return s.server.ListenAndServeTLS(s.config.Server.TLSCertPath, s.config.Server.TLSKeyPath)
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	// Stop the build queue
	if s.buildQueue != nil {
		log.Println("Stopping build queue...")
		s.buildQueue.Stop()
	}

	// Close notification service
	if s.notificationService != nil {
		log.Println("Stopping notification service...")
		s.notificationService.Close()
	}

	// Cancel server context
	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Close executor if it has a Close method
	if s.executor != nil {
		if closer, ok := s.executor.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("Warning: Failed to close executor: %v", err)
			}
		}
	}

	return s.server.Shutdown(shutdownCtx)
}

// Handler implementations
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.db.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status": "unhealthy", "error": "%s"}`, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "healthy"}`)
}

func (s *Server) handleAPIv1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "CASCI API v1", "version": "1.0.0"}`)
}

func (s *Server) handleJenkinsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{
		"mode": "NORMAL",
		"nodeDescription": "CASCI Server",
		"numExecutors": 2,
		"useSecurity": true,
		"jobs": []
	}`)
}

func (s *Server) handleJenkinsCrumb(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{
		"crumb": "casci-crumb-token",
		"crumbRequestField": "Jenkins-Crumb"
	}`)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket implementation will be added later
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprint(w, "WebSocket support coming soon")
}

// serveOpenAPIJSON serves the OpenAPI specification in JSON format
func (s *Server) serveOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(docs.SwaggerJSON)
}

// serveOpenAPIYAML serves the OpenAPI specification in YAML format
func (s *Server) serveOpenAPIYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(docs.SwaggerYAML)
}
