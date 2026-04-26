package graphql

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	
	"github.com/casapps/casci/src/pkg/users"
	"github.com/casapps/casci/src/pkg/projects"
	"github.com/casapps/casci/src/pkg/builds"
	"github.com/casapps/casci/src/pkg/nodes"
	"github.com/casapps/casci/src/pkg/metrics"
)

type Resolver struct{
	userService      *users.Service
	authManager      *users.AuthManager
	projectService   *projects.Service
	buildService     *builds.Service
	nodeService      *nodes.Service
	metricsCollector *metrics.Collector
}

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, input RegisterInput) (*AuthPayload, error) {
	panic("not implemented")
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input LoginInput) (*AuthPayload, error) {
	panic("not implemented")
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	panic("not implemented")
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// CreateProject is the resolver for the createProject field.
func (r *mutationResolver) CreateProject(ctx context.Context, input CreateProjectInput) (*Project, error) {
	panic("not implemented")
}

// UpdateProject is the resolver for the updateProject field.
func (r *mutationResolver) UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (*Project, error) {
	panic("not implemented")
}

// DeleteProject is the resolver for the deleteProject field.
func (r *mutationResolver) DeleteProject(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// TriggerBuild is the resolver for the triggerBuild field.
func (r *mutationResolver) TriggerBuild(ctx context.Context, projectID string, input *TriggerBuildInput) (*Build, error) {
	panic("not implemented")
}

// CancelBuild is the resolver for the cancelBuild field.
func (r *mutationResolver) CancelBuild(ctx context.Context, id string) (*Build, error) {
	panic("not implemented")
}

// RestartBuild is the resolver for the restartBuild field.
func (r *mutationResolver) RestartBuild(ctx context.Context, id string) (*Build, error) {
	panic("not implemented")
}

// RegisterNode is the resolver for the registerNode field.
func (r *mutationResolver) RegisterNode(ctx context.Context, input RegisterNodeInput) (*Node, error) {
	panic("not implemented")
}

// UpdateNode is the resolver for the updateNode field.
func (r *mutationResolver) UpdateNode(ctx context.Context, id string, input UpdateNodeInput) (*Node, error) {
	panic("not implemented")
}

// DeleteNode is the resolver for the deleteNode field.
func (r *mutationResolver) DeleteNode(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// DrainNode is the resolver for the drainNode field.
func (r *mutationResolver) DrainNode(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*User, error) {
	panic("not implemented")
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*User, error) {
	panic("not implemented")
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, limit *int, offset *int) (*UserConnection, error) {
	panic("not implemented")
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, id string) (*Project, error) {
	panic("not implemented")
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context, limit *int, offset *int) (*ProjectConnection, error) {
	panic("not implemented")
}

// Build is the resolver for the build field.
func (r *queryResolver) Build(ctx context.Context, id string) (*Build, error) {
	panic("not implemented")
}

// Builds is the resolver for the builds field.
func (r *queryResolver) Builds(ctx context.Context, projectID *string, limit *int, offset *int) (*BuildConnection, error) {
	panic("not implemented")
}

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id string) (*Node, error) {
	panic("not implemented")
}

// Nodes is the resolver for the nodes field.
func (r *queryResolver) Nodes(ctx context.Context, limit *int, offset *int) (*NodeConnection, error) {
	panic("not implemented")
}

// Health is the resolver for the health field.
func (r *queryResolver) Health(ctx context.Context) (*HealthStatus, error) {
	panic("not implemented")
}

// Metrics is the resolver for the metrics field.
func (r *queryResolver) Metrics(ctx context.Context) (*SystemMetrics, error) {
	panic("not implemented")
}

// BuildUpdated is the resolver for the buildUpdated field.
func (r *subscriptionResolver) BuildUpdated(ctx context.Context, projectID *string) (<-chan *Build, error) {
	panic("not implemented")
}

// BuildLogLine is the resolver for the buildLogLine field.
func (r *subscriptionResolver) BuildLogLine(ctx context.Context, buildID string) (<-chan *LogLine, error) {
	panic("not implemented")
}

// NodeStatusChanged is the resolver for the nodeStatusChanged field.
func (r *subscriptionResolver) NodeStatusChanged(ctx context.Context) (<-chan *Node, error) {
	panic("not implemented")
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
