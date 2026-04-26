package graphql

import (
	"github.com/casapps/casci/src/pkg/users"
	"github.com/casapps/casci/src/pkg/projects"
	"github.com/casapps/casci/src/pkg/builds"
	"github.com/casapps/casci/src/pkg/nodes"
	"github.com/casapps/casci/src/pkg/metrics"
)

// NewResolver creates a new GraphQL resolver with service dependencies
func NewResolver(
	userService *users.Service,
	authManager *users.AuthManager,
	projectService *projects.Service,
	buildService *builds.Service,
	nodeService *nodes.Service,
	metricsCollector *metrics.Collector,
) *Resolver {
	return &Resolver{
		userService:      userService,
		authManager:      authManager,
		projectService:   projectService,
		buildService:     buildService,
		nodeService:      nodeService,
		metricsCollector: metricsCollector,
	}
}
