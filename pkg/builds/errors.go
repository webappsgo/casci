package builds

import "errors"

var (
	// Database errors
	ErrBuildNotFound      = errors.New("build not found")
	ErrInvalidBuildStatus = errors.New("invalid build status")
	ErrInvalidTrigger     = errors.New("invalid build trigger")

	// Execution errors
	ErrBuildAlreadyRunning = errors.New("build already running")
	ErrBuildAlreadyQueued  = errors.New("build already queued")
	ErrCannotCancelBuild   = errors.New("cannot cancel completed build")
	ErrBuildFailed         = errors.New("build execution failed")

	// Authorization errors
	ErrUnauthorizedAccess = errors.New("unauthorized access to build")
)
