package git

import (
	"context"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Service handles Git operations
type Service struct{}

// NewService creates a new Git service
func NewService() *Service {
	return &Service{}
}

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	URL       string
	Branch    string
	CommitSHA string
	Depth     int
}

// Clone clones a repository to the specified path
func (s *Service) Clone(ctx context.Context, workspacePath string, opts CloneOptions) (*RepositoryInfo, error) {
	log.Printf("Cloning repository %s to %s", opts.URL, workspacePath)

	cloneOpts := &git.CloneOptions{
		URL:      opts.URL,
		Progress: nil, // Could add progress logging here
	}

	// Set branch if specified
	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		cloneOpts.SingleBranch = true
	}

	// Set depth for shallow clone
	if opts.Depth > 0 {
		cloneOpts.Depth = opts.Depth
	} else {
		cloneOpts.Depth = 1 // Default shallow clone
	}

	// Clone the repository
	repo, err := git.PlainCloneContext(ctx, workspacePath, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// If specific commit is requested, checkout that commit
	if opts.CommitSHA != "" {
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("failed to get worktree: %w", err)
		}

		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(opts.CommitSHA),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to checkout commit %s: %w", opts.CommitSHA, err)
		}
	}

	// Get repository information
	info, err := s.GetInfo(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	log.Printf("Successfully cloned repository: %s (commit: %s)", opts.URL, info.CommitSHA)
	return info, nil
}

// RepositoryInfo contains information about a cloned repository
type RepositoryInfo struct {
	CommitSHA     string
	CommitMessage string
	CommitAuthor  string
	Branch        string
}

// GetInfo extracts information from a repository
func (s *Service) GetInfo(repo *git.Repository) (*RepositoryInfo, error) {
	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get commit object
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	info := &RepositoryInfo{
		CommitSHA:     commit.Hash.String(),
		CommitMessage: commit.Message,
		CommitAuthor:  commit.Author.Name,
		Branch:        head.Name().Short(),
	}

	return info, nil
}

// ValidateURL checks if a Git URL is valid
func (s *Service) ValidateURL(url string) error {
	// Basic validation
	if url == "" {
		return fmt.Errorf("repository URL is empty")
	}

	// Check if it's a valid Git URL format
	// This is a simple check, could be more sophisticated
	if len(url) < 10 {
		return fmt.Errorf("repository URL is too short")
	}

	return nil
}
