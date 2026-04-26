package executor

import (
	"context"

	"github.com/casapps/casci/src/pkg/builds"
)

// Executor interface for build execution
type Executor interface {
	Execute(ctx context.Context, build *builds.Build) error
}

// ExecutorType represents the type of executor
type ExecutorType string

const (
	// TypeContainer uses Docker/Podman
	TypeContainer ExecutorType = "container"
	// TypeVM uses QEMU/KVM
	TypeVM ExecutorType = "vm"
	// TypeNative runs directly on host
	TypeNative ExecutorType = "native"
)

// Config holds executor configuration
type Config struct {
	Type           ExecutorType
	WorkspaceRoot  string
	CacheRoot      string
	ArtifactsRoot  string
	DefaultTimeout int // seconds
}
