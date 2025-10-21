package executor

import (
	"context"

	"github.com/casapps/casci/pkg/builds"
)

// Executor interface for build execution
type Executor interface {
	Execute(ctx context.Context, build *builds.Build) error
}

// ExecutorType represents the type of executor
type ExecutorType string

const (
	// ContainerExecutor uses Docker/Podman
	ContainerExecutor ExecutorType = "container"
	// VMExecutor uses QEMU/KVM
	VMExecutor ExecutorType = "vm"
	// NativeExecutor runs directly on host
	NativeExecutor ExecutorType = "native"
)

// Config holds executor configuration
type Config struct {
	Type           ExecutorType
	WorkspaceRoot  string
	CacheRoot      string
	ArtifactsRoot  string
	DefaultTimeout int // seconds
}
