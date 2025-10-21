package projects

import "errors"

var (
	// Validation errors
	ErrInvalidProjectName   = errors.New("invalid project name")
	ErrProjectNameTooShort  = errors.New("project name must be at least 3 characters")
	ErrInvalidRepositoryURL = errors.New("invalid repository URL")
	ErrInvalidBranch        = errors.New("invalid branch name")

	// Database errors
	ErrProjectNotFound   = errors.New("project not found")
	ErrProjectExists     = errors.New("project already exists")
	ErrProjectNameExists = errors.New("project name already exists for this user")

	// Authorization errors
	ErrUnauthorizedAccess = errors.New("unauthorized access to project")
	ErrNotProjectOwner    = errors.New("not project owner")
)
