package executor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"github.com/casapps/casci/src/pkg/builds"
	"github.com/casapps/casci/src/pkg/detection"
	"github.com/casapps/casci/src/pkg/git"
	"github.com/casapps/casci/src/pkg/security"
	"github.com/casapps/casci/src/pkg/workspace"
)

// ContainerExecutor executes builds in Docker containers
type ContainerExecutor struct {
	client          *client.Client
	workspace       *workspace.Manager
	git             *git.Service
	detector        *detection.Detector
	securityService *security.Service
	config          *Config
}

// NewContainerExecutor creates a new container executor
func NewContainerExecutor(config *Config, securityService *security.Service) (*ContainerExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &ContainerExecutor{
		client:          cli,
		workspace:       workspace.NewManager(config.WorkspaceRoot),
		git:             git.NewService(),
		detector:        detection.NewDetector(),
		securityService: securityService,
		config:          config,
	}, nil
}

// Execute runs a build in a container
func (ce *ContainerExecutor) Execute(ctx context.Context, build *builds.Build) error {
	log.Printf("Starting container execution for build #%d", build.BuildNumber)

	// Create workspace
	workspacePath, err := ce.workspace.Create(build.ProjectID, build.ID)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	defer ce.cleanup(build.ProjectID, build.ID)

	// Clone repository
	if err := ce.cloneRepository(ctx, build, workspacePath); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Auto-detect project type if not already configured
	var projectConfig *detection.ProjectConfig
	if build.ContainerImage == "" || build.BuildCommands == "" {
		log.Printf("Auto-detecting project type for build #%d", build.BuildNumber)
		var err error
		projectConfig, err = ce.detector.Detect(workspacePath)
		if err != nil {
			log.Printf("Warning: failed to detect project type: %v, using defaults", err)
		} else {
			log.Printf("Detected: %s %s (framework: %s, build tool: %s)",
				projectConfig.Type.Language,
				projectConfig.Type.LanguageVersion,
				projectConfig.Type.Framework,
				projectConfig.Type.BuildTool,
			)
		}
	}

	// Determine container image
	containerImage := ce.determineImage(build, projectConfig)

	// Pull image if needed
	if err := ce.pullImage(ctx, containerImage); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Create and run container
	if err := ce.runContainer(ctx, build, workspacePath, containerImage, projectConfig); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	// Run security scanning if service is available
	if ce.securityService != nil {
		log.Printf("Starting security scan for build #%d", build.BuildNumber)
		go func() {
			scanCtx := context.Background()
			result, err := ce.securityService.ScanBuild(scanCtx, build.ID, workspacePath)
			if err != nil {
				log.Printf("Security scan failed for build #%d: %v", build.BuildNumber, err)
				return
			}

			if !result.Summary.Passed {
				log.Printf("Security scan FAILED for build #%d: %d critical, %d high vulnerabilities, %d secrets",
					build.BuildNumber,
					result.Summary.CriticalCount,
					result.Summary.HighCount,
					result.Summary.SecretsFound,
				)
			} else {
				log.Printf("Security scan passed for build #%d", build.BuildNumber)
			}
		}()
	}

	log.Printf("Build #%d completed successfully", build.BuildNumber)
	return nil
}

// cloneRepository clones the repository into the workspace
func (ce *ContainerExecutor) cloneRepository(ctx context.Context, build *builds.Build, workspacePath string) error {
	log.Printf("Cloning repository for build #%d", build.BuildNumber)

	cloneOpts := git.CloneOptions{
		URL:       build.RepositoryURL,
		Branch:    build.Branch,
		CommitSHA: build.CommitSHA,
		Depth:     1, // Shallow clone
	}

	_, err := ce.git.Clone(ctx, workspacePath, cloneOpts)
	return err
}

// determineImage determines which container image to use
func (ce *ContainerExecutor) determineImage(build *builds.Build, projectConfig *detection.ProjectConfig) string {
	// Use explicitly configured image first
	if build.ContainerImage != "" {
		return build.ContainerImage
	}

	// Use detected image
	if projectConfig != nil && projectConfig.ContainerImage != "" {
		return projectConfig.ContainerImage
	}

	// Default fallback
	return "ubuntu:22.04"
}

// pullImage pulls the container image if not present
func (ce *ContainerExecutor) pullImage(ctx context.Context, imageName string) error {
	log.Printf("Pulling image: %s", imageName)

	reader, err := ce.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Read pull output to ensure it completes
	_, err = io.Copy(io.Discard, reader)
	return err
}

// runContainer creates and runs the build container
func (ce *ContainerExecutor) runContainer(ctx context.Context, build *builds.Build, workspacePath, imageName string, projectConfig *detection.ProjectConfig) error {
	log.Printf("Creating container for build #%d", build.BuildNumber)

	// Container configuration
	containerConfig := &container.Config{
		Image:      imageName,
		WorkingDir: "/workspace",
		Cmd:        ce.buildCommands(build, projectConfig),
		Env:        ce.buildEnvironment(build, projectConfig),
	}

	// Host configuration with mounts and resource limits
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workspacePath,
				Target: "/workspace",
			},
		},
		Resources: container.Resources{
			Memory:   2 * 1024 * 1024 * 1024, // 2GB default
			NanoCPUs: 2 * 1000000000,         // 2 CPUs default
		},
		AutoRemove: true,
	}

	// Create container
	resp, err := ce.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID
	log.Printf("Created container %s for build #%d", containerID[:12], build.BuildNumber)

	// Start container
	if err := ce.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Stream logs
	if err := ce.streamLogs(ctx, containerID, build); err != nil {
		log.Printf("Warning: failed to stream logs: %v", err)
	}

	// Wait for container to finish
	statusCh, errCh := ce.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("container wait error: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with status %d", status.StatusCode)
		}
	}

	return nil
}

// buildCommands generates the command to run in the container
func (ce *ContainerExecutor) buildCommands(build *builds.Build, projectConfig *detection.ProjectConfig) []string {
	// Use explicitly configured commands first
	if build.BuildCommands != "" {
		return []string{"/bin/sh", "-c", build.BuildCommands}
	}

	// Use parsed pipeline commands
	if projectConfig != nil && projectConfig.Pipeline != nil {
		commands := ce.extractPipelineCommands(projectConfig.Pipeline)
		if len(commands) > 0 {
			commandStr := strings.Join(commands, " && ")
			log.Printf("Using parsed pipeline commands from %s: %s", projectConfig.PipelineFormat, commandStr)
			return []string{"/bin/sh", "-c", commandStr}
		}
	}

	// Use detected commands
	if projectConfig != nil && len(projectConfig.BuildCommands) > 0 {
		commands := strings.Join(projectConfig.BuildCommands, " && ")
		log.Printf("Using detected build commands: %s", commands)
		return []string{"/bin/sh", "-c", commands}
	}

	// Default build script
	log.Printf("No build commands configured, using default inspection script")
	return []string{
		"/bin/sh",
		"-c",
		`echo "Build started"
echo "Working directory: $(pwd)"
echo "Repository contents:"
ls -la
echo "Build completed successfully"`,
	}
}

// extractPipelineCommands extracts all commands from a parsed pipeline
func (ce *ContainerExecutor) extractPipelineCommands(p interface{}) []string {
	var commands []string

	// Type assertion for pipeline.Pipeline
	type pipelineStage struct {
		Name     string
		Parallel bool
		Jobs     []struct {
			Name        string
			Container   string
			Commands    []string
			Environment map[string]string
		}
	}

	type pipelineType struct {
		Stages []pipelineStage
	}

	// Extract commands from all stages and jobs
	if pipeline, ok := p.(*pipelineType); ok {
		for _, stage := range pipeline.Stages {
			for _, job := range stage.Jobs {
				commands = append(commands, job.Commands...)
			}
		}
	}

	return commands
}

// buildEnvironment generates environment variables for the container
func (ce *ContainerExecutor) buildEnvironment(build *builds.Build, projectConfig *detection.ProjectConfig) []string {
	env := []string{
		fmt.Sprintf("CI=true"),
		fmt.Sprintf("CASCI=true"),
		fmt.Sprintf("BUILD_ID=%d", build.ID),
		fmt.Sprintf("BUILD_NUMBER=%d", build.BuildNumber),
		fmt.Sprintf("PROJECT_ID=%d", build.ProjectID),
		fmt.Sprintf("COMMIT_SHA=%s", build.CommitSHA),
		fmt.Sprintf("BRANCH=%s", build.Branch),
	}

	// Add detected environment variables
	if projectConfig != nil && projectConfig.Environment != nil {
		for key, value := range projectConfig.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// TODO: Add user-defined environment variables and secrets
	return env
}

// streamLogs streams container logs to a file
func (ce *ContainerExecutor) streamLogs(ctx context.Context, containerID string, build *builds.Build) error {
	// Create log file
	logPath := filepath.Join("/var/log/casci/builds", fmt.Sprintf("build-%d.log", build.ID))
	os.MkdirAll(filepath.Dir(logPath), 0755)

	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// Stream logs from container
	reader, err := ce.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Copy logs to file
	_, err = io.Copy(logFile, reader)
	return err
}

// cleanup removes the workspace after build
func (ce *ContainerExecutor) cleanup(projectID, buildID int) {
	log.Printf("Cleaning up workspace for project %d, build %d", projectID, buildID)
	if err := ce.workspace.Cleanup(projectID, buildID); err != nil {
		log.Printf("Warning: failed to cleanup workspace: %v", err)
	}
}

// Close closes the Docker client
func (ce *ContainerExecutor) Close() error {
	if ce.client != nil {
		return ce.client.Close()
	}
	return nil
}

// HealthCheck checks if Docker is available
func (ce *ContainerExecutor) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := ce.client.Ping(ctx)
	return err
}
