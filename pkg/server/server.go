package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/casapps/casci/internal/config"
	"github.com/casapps/casci/pkg/builds"
	"github.com/casapps/casci/pkg/credentials"
	"github.com/casapps/casci/pkg/database"
	"github.com/casapps/casci/pkg/executor"
	"github.com/casapps/casci/pkg/metrics"
	"github.com/casapps/casci/pkg/nodes"
	"github.com/casapps/casci/pkg/notifications"
	"github.com/casapps/casci/pkg/projects"
	"github.com/casapps/casci/pkg/queue"
	"github.com/casapps/casci/pkg/security"
	"github.com/casapps/casci/pkg/users"
	"github.com/casapps/casci/pkg/webhooks"
)

// Server represents the HTTP/WebSocket server
type Server struct {
	config              *config.Config
	db                  *database.Database
	router              *http.ServeMux
	server              *http.Server
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
	webhookHandler      *webhooks.Handler
	metricsCollector    *metrics.Collector
	metricsExporter     *metrics.PrometheusExporter
	metricsHandler      *metrics.Handler
	executor            executor.Executor
	buildQueue          *queue.BuildQueue
	ctx                 context.Context
	cancelFunc          context.CancelFunc
}

// New creates a new server instance
func New(cfg *config.Config, db *database.Database) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	srv := &Server{
		config:     cfg,
		db:         db,
		router:     http.NewServeMux(),
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// Initialize user management
	srv.userService = users.NewService(db)
	srv.authManager = users.NewAuthManager()
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
	securityRepo := security.NewSQLRepository(db.DB)
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
	notificationRepo := notifications.NewSQLRepository(db.DB)
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
	credRepo := credentials.NewRepository(db.DB)
	srv.credentialService = credentials.NewService(credRepo, cfg.EncryptionKey)
	srv.credentialHandler = credentials.NewHandler(srv.credentialService)
	log.Printf("Credential management initialized")

	// Initialize webhook handler
	srv.webhookHandler = webhooks.NewHandler(srv.projectService, srv.buildService)

	// Initialize executor
	executorConfig := &executor.Config{
		Type:           executor.ContainerExecutor,
		WorkspaceRoot:  "/var/lib/casci/workspaces",
		CacheRoot:      "/var/lib/casci/cache",
		ArtifactsRoot:  "/var/lib/casci/artifacts",
		DefaultTimeout: 3600, // 1 hour default
	}

	var err error
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

	// Create HTTP server
	srv.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: srv.router,
		// Note: No timeouts as per spec requirements
	}

	return srv, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check and metrics endpoints (public)
	s.router.HandleFunc("/health", s.metricsHandler.ServeHealth)
	s.router.HandleFunc("/healthz", s.metricsHandler.ServeHealth)
	s.router.HandleFunc("/readyz", s.metricsHandler.ServeReadiness)
	s.router.HandleFunc("/livez", s.metricsHandler.ServeLiveness)
	s.router.HandleFunc("/metrics", s.metricsHandler.ServeMetrics)
	s.router.HandleFunc("/metrics/json", s.metricsHandler.ServeMetricsJSON)

	// Detailed metrics endpoints (public - for monitoring systems)
	s.router.HandleFunc("/api/v1/metrics/system", s.metricsHandler.ServeSystemMetrics)
	s.router.HandleFunc("/api/v1/metrics/builds", s.metricsHandler.ServeBuildMetrics)
	s.router.HandleFunc("/api/v1/metrics/nodes", s.metricsHandler.ServeNodeMetrics)
	s.router.HandleFunc("/api/v1/metrics/security", s.metricsHandler.ServeSecurityMetrics)
	s.router.HandleFunc("/api/v1/metrics/api", s.metricsHandler.ServeAPIMetrics)

	// Authentication endpoints (public)
	s.router.HandleFunc("/api/v1/auth/register", s.userHandler.Register)
	s.router.HandleFunc("/api/v1/auth/login", s.userHandler.Login)
	s.router.Handle("/api/v1/auth/refresh", s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.RefreshToken)))

	// User endpoints (authenticated)
	s.router.Handle("/api/v1/users/me", s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.GetMe)))
	s.router.Handle("/api/v1/users/me/token", s.userMiddleware.RequireAuth(http.HandlerFunc(s.userHandler.RegenerateAPIToken)))

	// User management (admin only)
	s.router.Handle("/api/v1/users", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.userHandler.ListUsers)))

	// Individual user operations
	s.router.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
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
	s.router.Handle("/api/v1/projects", s.userMiddleware.RequireAuth(http.HandlerFunc(s.projectHandler.List)))
	s.router.HandleFunc("/api/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on method
		switch r.Method {
		case http.MethodPost:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.projectHandler.Create)).ServeHTTP(w, r)
		case http.MethodGet:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.projectHandler.Get)).ServeHTTP(w, r)
		case http.MethodPut, http.MethodPatch:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.projectHandler.Update)).ServeHTTP(w, r)
		case http.MethodDelete:
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.projectHandler.Delete)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Build endpoints for a project (authenticated)
	s.router.Handle("/api/v1/projects/", s.userMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle /api/v1/projects/{id}/builds routes
		if r.URL.Path == "/api/v1/projects/" || !contains(r.URL.Path, "/builds") {
			return // Not a builds route
		}

		if contains(r.URL.Path, "/builds/stats") {
			s.buildHandler.GetStats(w, r)
		} else if r.Method == http.MethodPost && !contains(r.URL.Path, "/builds/") {
			// POST /api/v1/projects/{id}/builds - Trigger build
			s.buildHandler.Trigger(w, r)
		} else if r.Method == http.MethodGet {
			// GET /api/v1/projects/{id}/builds - List builds
			s.buildHandler.ListByProject(w, r)
		}
	})))

	// Individual build endpoints (authenticated)
	s.router.Handle("/api/v1/builds/", s.userMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "/log") {
			s.buildHandler.GetLog(w, r)
		} else if contains(r.URL.Path, "/cancel") {
			s.buildHandler.Cancel(w, r)
		} else if contains(r.URL.Path, "/restart") {
			s.buildHandler.Restart(w, r)
		} else if r.Method == http.MethodGet {
			s.buildHandler.Get(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Node endpoints (authenticated)
	s.router.Handle("/api/v1/nodes", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.List)))
	s.router.HandleFunc("/api/v1/nodes/", func(w http.ResponseWriter, r *http.Request) {
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
	s.router.HandleFunc("/api/v1/nodes/register", s.nodeHandler.Register)

	// Node token generation (admin only)
	s.router.Handle("/api/v1/nodes/token", s.userMiddleware.RequireAdmin(http.HandlerFunc(s.nodeHandler.GenerateToken)))

	// Security endpoints (authenticated)
	s.router.HandleFunc("/api/v1/builds/", func(w http.ResponseWriter, r *http.Request) {
		// Handle security routes for builds
		if contains(r.URL.Path, "/security") && !contains(r.URL.Path, "/security/scan") {
			// GET /api/v1/builds/{id}/security - Get security reports
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.GetBuildSecurityReports)).ServeHTTP(w, r)
		} else if contains(r.URL.Path, "/security/scan") && r.Method == http.MethodPost {
			// POST /api/v1/builds/{id}/security/scan - Trigger scan
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.TriggerScan)).ServeHTTP(w, r)
		}
	})

	s.router.Handle("/api/v1/security/reports", s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.ListSecurityReports)))
	s.router.HandleFunc("/api/v1/security/reports/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.GetSecurityReport)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	s.router.Handle("/api/v1/security/statistics", s.userMiddleware.RequireAuth(http.HandlerFunc(s.securityHandler.GetStatistics)))
	s.router.Handle("/api/v1/security/config", s.userMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.securityHandler.GetConfig(w, r)
		} else if r.Method == http.MethodPut {
			s.securityHandler.UpdateConfig(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Notification endpoints (authenticated)
	s.router.Handle("/api/v1/notifications", s.userMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.notificationHandler.ListConfigs(w, r)
		} else if r.Method == http.MethodPost {
			s.notificationHandler.CreateConfig(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	s.router.HandleFunc("/api/v1/notifications/", func(w http.ResponseWriter, r *http.Request) {
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
	s.router.Handle("/api/v1/notifications/types", s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetSupportedTypes)))
	s.router.Handle("/api/v1/notifications/events", s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetEventTypes)))

	// Project notification configs
	s.router.HandleFunc("/api/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "/notifications") && r.Method == http.MethodGet {
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetProjectConfigs)).ServeHTTP(w, r)
		}
	})

	// Build notification logs
	s.router.HandleFunc("/api/v1/builds/", func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "/notifications") && r.Method == http.MethodGet {
			s.userMiddleware.RequireAuth(http.HandlerFunc(s.notificationHandler.GetBuildLogs)).ServeHTTP(w, r)
		}
	})

	// User credential endpoints (authenticated)
	s.router.Handle("/api/v1/credentials/user", s.userMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.credentialHandler.ListUserCredentials(w, r)
		} else if r.Method == http.MethodPost {
			s.credentialHandler.CreateUserCredential(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	s.router.HandleFunc("/api/v1/credentials/user/", func(w http.ResponseWriter, r *http.Request) {
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

	// Project credential endpoints (authenticated)
	s.router.HandleFunc("/api/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "/credentials") {
			if r.Method == http.MethodGet {
				s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.ListProjectCredentials)).ServeHTTP(w, r)
			} else if r.Method == http.MethodPost {
				s.userMiddleware.RequireAuth(http.HandlerFunc(s.credentialHandler.CreateProjectCredential)).ServeHTTP(w, r)
			}
		}
	})

	s.router.HandleFunc("/api/v1/credentials/project/", func(w http.ResponseWriter, r *http.Request) {
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
	s.router.HandleFunc("/api/v1/", s.handleAPIv1)

	// Jenkins compatibility routes
	s.router.HandleFunc("/api/json", s.handleJenkinsAPI)
	s.router.HandleFunc("/crumbIssuer/api/json", s.handleJenkinsCrumb)

	// Webhook endpoint (public - secured by provider signatures)
	s.router.HandleFunc("/webhook", s.webhookHandler.ServeHTTP)

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.handleWebSocket)

	// Static files / Web UI (will be implemented later)
	s.router.HandleFunc("/", s.handleWebUI)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && anyIndexOf(s, substr) >= 0
}

func anyIndexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s:%d", s.config.Host, s.config.Port)

	if s.config.TLSEnabled {
		return s.server.ListenAndServeTLS(s.config.TLSCertPath, s.config.TLSKeyPath)
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

func (s *Server) handleWebUI(w http.ResponseWriter, r *http.Request) {
	// Serve embedded UI (will be implemented later)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
	<title>CASCI</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
			margin: 0;
			padding: 0;
			display: flex;
			justify-content: center;
			align-items: center;
			min-height: 100vh;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			color: white;
		}
		.container {
			text-align: center;
			padding: 2rem;
		}
		h1 {
			font-size: 4rem;
			margin: 0;
		}
		p {
			font-size: 1.5rem;
			margin: 1rem 0;
		}
		.status {
			background: rgba(255,255,255,0.2);
			padding: 1rem;
			border-radius: 8px;
			margin-top: 2rem;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>🚀 CASCI</h1>
		<p>CI/CD Application Server</p>
		<div class="status">
			<p>✅ Server is running</p>
			<p>Web UI coming soon...</p>
		</div>
	</div>
</body>
</html>`)
}
