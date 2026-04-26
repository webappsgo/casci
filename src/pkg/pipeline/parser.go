package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PipelineFormat represents the detected CI/CD format
type PipelineFormat string

const (
	FormatJenkinsfile      PipelineFormat = "jenkinsfile"
	FormatGitHubActions    PipelineFormat = "github-actions"
	FormatGitLabCI         PipelineFormat = "gitlab-ci"
	FormatCircleCI         PipelineFormat = "circleci"
	FormatTravisCI         PipelineFormat = "travis-ci"
	FormatAzurePipelines   PipelineFormat = "azure-pipelines"
	FormatBitbucketPipelines PipelineFormat = "bitbucket-pipelines"
	FormatDroneCI          PipelineFormat = "drone-ci"
	FormatBuildkite        PipelineFormat = "buildkite"
	FormatUnknown          PipelineFormat = "unknown"
)

// Pipeline represents the internal pipeline format
type Pipeline struct {
	Version     string                 `json:"version"`
	Environment map[string]string      `json:"environment"`
	Secrets     []string               `json:"secrets"`
	Stages      []Stage                `json:"stages"`
}

// Stage represents a pipeline stage
type Stage struct {
	Name      string   `json:"name"`
	Parallel  bool     `json:"parallel"`
	Jobs      []Job    `json:"jobs"`
	When      string   `json:"when,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
}

// Job represents a job within a stage
type Job struct {
	Name        string            `json:"name"`
	Container   string            `json:"container,omitempty"`
	Commands    []string          `json:"commands"`
	Environment map[string]string `json:"environment,omitempty"`
	Artifacts   []Artifact        `json:"artifacts,omitempty"`
	Cache       []string          `json:"cache,omitempty"`
	Services    []Service         `json:"services,omitempty"`
}

// Artifact represents a build artifact
type Artifact struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// Service represents a sidecar service
type Service struct {
	Name  string            `json:"name"`
	Image string            `json:"image"`
	Env   map[string]string `json:"env,omitempty"`
}

// Parser handles pipeline detection and parsing
type Parser struct{}

// NewParser creates a new pipeline parser
func NewParser() *Parser {
	return &Parser{}
}

// DetectFormat detects the CI/CD format in a repository
func (p *Parser) DetectFormat(repoPath string) PipelineFormat {
	checks := []struct {
		path   string
		format PipelineFormat
	}{
		{"Jenkinsfile", FormatJenkinsfile},
		{".github/workflows", FormatGitHubActions},
		{".gitlab-ci.yml", FormatGitLabCI},
		{".circleci/config.yml", FormatCircleCI},
		{".travis.yml", FormatTravisCI},
		{"azure-pipelines.yml", FormatAzurePipelines},
		{"bitbucket-pipelines.yml", FormatBitbucketPipelines},
		{".drone.yml", FormatDroneCI},
		{"buildkite.yml", FormatBuildkite},
	}

	for _, check := range checks {
		fullPath := filepath.Join(repoPath, check.path)
		if _, err := os.Stat(fullPath); err == nil {
			return check.format
		}
	}

	return FormatUnknown
}

// Parse parses a pipeline file and converts it to the internal format
func (p *Parser) Parse(repoPath string) (*Pipeline, error) {
	format := p.DetectFormat(repoPath)

	switch format {
	case FormatJenkinsfile:
		return p.parseJenkinsfile(repoPath)
	case FormatGitHubActions:
		return p.parseGitHubActions(repoPath)
	case FormatGitLabCI:
		return p.parseGitLabCI(repoPath)
	case FormatCircleCI:
		return p.parseCircleCI(repoPath)
	case FormatTravisCI:
		return p.parseTravisCI(repoPath)
	case FormatAzurePipelines:
		return p.parseAzurePipelines(repoPath)
	case FormatBitbucketPipelines:
		return p.parseBitbucketPipelines(repoPath)
	case FormatDroneCI:
		return p.parseDroneCI(repoPath)
	case FormatBuildkite:
		return p.parseBuildkite(repoPath)
	default:
		return nil, fmt.Errorf("no supported pipeline format detected")
	}
}

// parseJenkinsfile parses a Jenkinsfile (basic support)
func (p *Parser) parseJenkinsfile(repoPath string) (*Pipeline, error) {
	jenkinsfilePath := filepath.Join(repoPath, "Jenkinsfile")
	data, err := os.ReadFile(jenkinsfilePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Basic Jenkinsfile parsing (declarative pipeline)
	if strings.Contains(content, "pipeline {") {
		// Extract stages
		if strings.Contains(content, "stages {") {
			// Simple stage extraction
			stage := Stage{
				Name:     "build",
				Parallel: false,
				Jobs: []Job{
					{
						Name:     "jenkinsfile-build",
						Commands: []string{"echo 'Running Jenkinsfile pipeline'"},
					},
				},
			}
			pipeline.Stages = append(pipeline.Stages, stage)
		}
	}

	return pipeline, nil
}

// parseGitHubActions parses GitHub Actions workflows
func (p *Parser) parseGitHubActions(repoPath string) (*Pipeline, error) {
	workflowsDir := filepath.Join(repoPath, ".github", "workflows")
	files, err := os.ReadDir(workflowsDir)
	if err != nil {
		return nil, err
	}

	// Parse the first workflow file
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml")) {
			return p.parseGitHubActionsFile(filepath.Join(workflowsDir, file.Name()))
		}
	}

	return nil, fmt.Errorf("no workflow files found")
}

func (p *Parser) parseGitHubActionsFile(filePath string) (*Pipeline, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var ghWorkflow map[string]interface{}
	if err := yaml.Unmarshal(data, &ghWorkflow); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract jobs
	if jobs, ok := ghWorkflow["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			job := p.convertGitHubActionsJob(jobName, jobData)
			stage := Stage{
				Name:     jobName,
				Parallel: false,
				Jobs:     []Job{job},
			}
			pipeline.Stages = append(pipeline.Stages, stage)
		}
	}

	return pipeline, nil
}

func (p *Parser) convertGitHubActionsJob(name string, data interface{}) Job {
	job := Job{
		Name:        name,
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	jobMap, ok := data.(map[string]interface{})
	if !ok {
		return job
	}

	// Extract container
	if runsOn, ok := jobMap["runs-on"].(string); ok {
		job.Container = p.mapGitHubRunsOnToContainer(runsOn)
	}

	// Extract steps
	if steps, ok := jobMap["steps"].([]interface{}); ok {
		for _, step := range steps {
			if stepMap, ok := step.(map[string]interface{}); ok {
				if run, ok := stepMap["run"].(string); ok {
					job.Commands = append(job.Commands, run)
				}
			}
		}
	}

	return job
}

func (p *Parser) mapGitHubRunsOnToContainer(runsOn string) string {
	mapping := map[string]string{
		"ubuntu-latest":   "ubuntu:22.04",
		"ubuntu-22.04":    "ubuntu:22.04",
		"ubuntu-20.04":    "ubuntu:20.04",
		"macos-latest":    "",
		"windows-latest":  "",
	}

	if container, ok := mapping[runsOn]; ok {
		return container
	}
	return "ubuntu:22.04"
}

// parseGitLabCI parses GitLab CI configuration
func (p *Parser) parseGitLabCI(repoPath string) (*Pipeline, error) {
	gitlabCIPath := filepath.Join(repoPath, ".gitlab-ci.yml")
	data, err := os.ReadFile(gitlabCIPath)
	if err != nil {
		return nil, err
	}

	var gitlabCI map[string]interface{}
	if err := yaml.Unmarshal(data, &gitlabCI); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract stages
	stagesRaw, hasStages := gitlabCI["stages"]
	if hasStages {
		if stagesList, ok := stagesRaw.([]interface{}); ok {
			for _, stageName := range stagesList {
				if name, ok := stageName.(string); ok {
					stage := Stage{
						Name:     name,
						Parallel: true,
						Jobs:     []Job{},
					}
					pipeline.Stages = append(pipeline.Stages, stage)
				}
			}
		}
	}

	// Extract jobs
	for key, value := range gitlabCI {
		if key == "stages" || key == "variables" || key == "image" || strings.HasPrefix(key, ".") {
			continue
		}

		if jobData, ok := value.(map[string]interface{}); ok {
			job := p.convertGitLabCIJob(key, jobData)

			// Add to appropriate stage
			stageName := "build"
			if stg, ok := jobData["stage"].(string); ok {
				stageName = stg
			}

			// Find or create stage
			stageIndex := -1
			for i, stage := range pipeline.Stages {
				if stage.Name == stageName {
					stageIndex = i
					break
				}
			}

			if stageIndex == -1 {
				pipeline.Stages = append(pipeline.Stages, Stage{
					Name:     stageName,
					Parallel: true,
					Jobs:     []Job{job},
				})
			} else {
				pipeline.Stages[stageIndex].Jobs = append(pipeline.Stages[stageIndex].Jobs, job)
			}
		}
	}

	return pipeline, nil
}

func (p *Parser) convertGitLabCIJob(name string, data map[string]interface{}) Job {
	job := Job{
		Name:        name,
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	// Extract image
	if image, ok := data["image"].(string); ok {
		job.Container = image
	}

	// Extract script
	if script, ok := data["script"].([]interface{}); ok {
		for _, cmd := range script {
			if cmdStr, ok := cmd.(string); ok {
				job.Commands = append(job.Commands, cmdStr)
			}
		}
	}

	return job
}

// parseCircleCI parses CircleCI configuration
func (p *Parser) parseCircleCI(repoPath string) (*Pipeline, error) {
	circleCIPath := filepath.Join(repoPath, ".circleci", "config.yml")
	data, err := os.ReadFile(circleCIPath)
	if err != nil {
		return nil, err
	}

	var circleCI map[string]interface{}
	if err := yaml.Unmarshal(data, &circleCI); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract jobs
	if jobs, ok := circleCI["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			job := p.convertCircleCIJob(jobName, jobData)
			stage := Stage{
				Name:     jobName,
				Parallel: false,
				Jobs:     []Job{job},
			}
			pipeline.Stages = append(pipeline.Stages, stage)
		}
	}

	return pipeline, nil
}

func (p *Parser) convertCircleCIJob(name string, data interface{}) Job {
	job := Job{
		Name:        name,
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	jobMap, ok := data.(map[string]interface{})
	if !ok {
		return job
	}

	// Extract docker image
	if docker, ok := jobMap["docker"].([]interface{}); ok && len(docker) > 0 {
		if dockerMap, ok := docker[0].(map[string]interface{}); ok {
			if image, ok := dockerMap["image"].(string); ok {
				job.Container = image
			}
		}
	}

	// Extract steps
	if steps, ok := jobMap["steps"].([]interface{}); ok {
		for _, step := range steps {
			if stepMap, ok := step.(map[string]interface{}); ok {
				if run, ok := stepMap["run"].(string); ok {
					job.Commands = append(job.Commands, run)
				} else if runMap, ok := stepMap["run"].(map[string]interface{}); ok {
					if cmd, ok := runMap["command"].(string); ok {
						job.Commands = append(job.Commands, cmd)
					}
				}
			}
		}
	}

	return job
}

// parseTravisCI parses Travis CI configuration
func (p *Parser) parseTravisCI(repoPath string) (*Pipeline, error) {
	travisPath := filepath.Join(repoPath, ".travis.yml")
	data, err := os.ReadFile(travisPath)
	if err != nil {
		return nil, err
	}

	var travisCI map[string]interface{}
	if err := yaml.Unmarshal(data, &travisCI); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Build job
	job := Job{
		Name:        "build",
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	// Extract language for container
	if lang, ok := travisCI["language"].(string); ok {
		job.Container = p.mapTravisLanguageToContainer(lang)
	}

	// Extract script
	if script, ok := travisCI["script"].([]interface{}); ok {
		for _, cmd := range script {
			if cmdStr, ok := cmd.(string); ok {
				job.Commands = append(job.Commands, cmdStr)
			}
		}
	} else if script, ok := travisCI["script"].(string); ok {
		job.Commands = append(job.Commands, script)
	}

	stage := Stage{
		Name:     "build",
		Parallel: false,
		Jobs:     []Job{job},
	}
	pipeline.Stages = append(pipeline.Stages, stage)

	return pipeline, nil
}

func (p *Parser) mapTravisLanguageToContainer(lang string) string {
	mapping := map[string]string{
		"node_js": "node:18-alpine",
		"python":  "python:3.11-slim",
		"ruby":    "ruby:3.2-alpine",
		"go":      "golang:1.21-alpine",
		"java":    "maven:3.9-eclipse-temurin-17",
	}

	if container, ok := mapping[lang]; ok {
		return container
	}
	return "ubuntu:22.04"
}

// parseAzurePipelines parses Azure Pipelines configuration
func (p *Parser) parseAzurePipelines(repoPath string) (*Pipeline, error) {
	azurePath := filepath.Join(repoPath, "azure-pipelines.yml")
	data, err := os.ReadFile(azurePath)
	if err != nil {
		return nil, err
	}

	var azureCI map[string]interface{}
	if err := yaml.Unmarshal(data, &azureCI); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract stages
	if stages, ok := azureCI["stages"].([]interface{}); ok {
		for _, stageData := range stages {
			if stageMap, ok := stageData.(map[string]interface{}); ok {
				stage := p.convertAzureStage(stageMap)
				pipeline.Stages = append(pipeline.Stages, stage)
			}
		}
	}

	return pipeline, nil
}

func (p *Parser) convertAzureStage(data map[string]interface{}) Stage {
	stage := Stage{
		Name:     "build",
		Parallel: false,
		Jobs:     []Job{},
	}

	if name, ok := data["stage"].(string); ok {
		stage.Name = name
	}

	if jobs, ok := data["jobs"].([]interface{}); ok {
		for _, jobData := range jobs {
			if jobMap, ok := jobData.(map[string]interface{}); ok {
				job := p.convertAzureJob(jobMap)
				stage.Jobs = append(stage.Jobs, job)
			}
		}
	}

	return stage
}

func (p *Parser) convertAzureJob(data map[string]interface{}) Job {
	job := Job{
		Name:        "job",
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	if name, ok := data["job"].(string); ok {
		job.Name = name
	}

	if steps, ok := data["steps"].([]interface{}); ok {
		for _, step := range steps {
			if stepMap, ok := step.(map[string]interface{}); ok {
				if script, ok := stepMap["script"].(string); ok {
					job.Commands = append(job.Commands, script)
				}
			}
		}
	}

	return job
}

// parseBitbucketPipelines parses Bitbucket Pipelines configuration
func (p *Parser) parseBitbucketPipelines(repoPath string) (*Pipeline, error) {
	bitbucketPath := filepath.Join(repoPath, "bitbucket-pipelines.yml")
	data, err := os.ReadFile(bitbucketPath)
	if err != nil {
		return nil, err
	}

	var bitbucket map[string]interface{}
	if err := yaml.Unmarshal(data, &bitbucket); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract default pipeline
	if pipelines, ok := bitbucket["pipelines"].(map[string]interface{}); ok {
		if defaultPipeline, ok := pipelines["default"].([]interface{}); ok {
			for i, stepData := range defaultPipeline {
				if stepMap, ok := stepData.(map[string]interface{}); ok {
					job := p.convertBitbucketStep(stepMap)
					stage := Stage{
						Name:     fmt.Sprintf("step-%d", i+1),
						Parallel: false,
						Jobs:     []Job{job},
					}
					pipeline.Stages = append(pipeline.Stages, stage)
				}
			}
		}
	}

	return pipeline, nil
}

func (p *Parser) convertBitbucketStep(data map[string]interface{}) Job {
	job := Job{
		Name:        "step",
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	if step, ok := data["step"].(map[string]interface{}); ok {
		if name, ok := step["name"].(string); ok {
			job.Name = name
		}

		if script, ok := step["script"].([]interface{}); ok {
			for _, cmd := range script {
				if cmdStr, ok := cmd.(string); ok {
					job.Commands = append(job.Commands, cmdStr)
				}
			}
		}
	}

	return job
}

// parseDroneCI parses Drone CI configuration
func (p *Parser) parseDroneCI(repoPath string) (*Pipeline, error) {
	dronePath := filepath.Join(repoPath, ".drone.yml")
	data, err := os.ReadFile(dronePath)
	if err != nil {
		return nil, err
	}

	var drone map[string]interface{}
	if err := yaml.Unmarshal(data, &drone); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract steps
	if steps, ok := drone["steps"].([]interface{}); ok {
		for _, stepData := range steps {
			if stepMap, ok := stepData.(map[string]interface{}); ok {
				job := p.convertDroneStep(stepMap)
				stage := Stage{
					Name:     job.Name,
					Parallel: false,
					Jobs:     []Job{job},
				}
				pipeline.Stages = append(pipeline.Stages, stage)
			}
		}
	}

	return pipeline, nil
}

func (p *Parser) convertDroneStep(data map[string]interface{}) Job {
	job := Job{
		Name:        "step",
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	if name, ok := data["name"].(string); ok {
		job.Name = name
	}

	if image, ok := data["image"].(string); ok {
		job.Container = image
	}

	if commands, ok := data["commands"].([]interface{}); ok {
		for _, cmd := range commands {
			if cmdStr, ok := cmd.(string); ok {
				job.Commands = append(job.Commands, cmdStr)
			}
		}
	}

	return job
}

// parseBuildkite parses Buildkite configuration
func (p *Parser) parseBuildkite(repoPath string) (*Pipeline, error) {
	buildkitePath := filepath.Join(repoPath, "buildkite.yml")
	data, err := os.ReadFile(buildkitePath)
	if err != nil {
		// Try .buildkite/pipeline.yml
		buildkitePath = filepath.Join(repoPath, ".buildkite", "pipeline.yml")
		data, err = os.ReadFile(buildkitePath)
		if err != nil {
			return nil, err
		}
	}

	var buildkite map[string]interface{}
	if err := yaml.Unmarshal(data, &buildkite); err != nil {
		return nil, err
	}

	pipeline := &Pipeline{
		Version:     "1.0",
		Environment: make(map[string]string),
		Stages:      []Stage{},
	}

	// Extract steps
	if steps, ok := buildkite["steps"].([]interface{}); ok {
		for _, stepData := range steps {
			if stepMap, ok := stepData.(map[string]interface{}); ok {
				job := p.convertBuildkiteStep(stepMap)
				stage := Stage{
					Name:     job.Name,
					Parallel: false,
					Jobs:     []Job{job},
				}
				pipeline.Stages = append(pipeline.Stages, stage)
			}
		}
	}

	return pipeline, nil
}

func (p *Parser) convertBuildkiteStep(data map[string]interface{}) Job {
	job := Job{
		Name:        "step",
		Commands:    []string{},
		Environment: make(map[string]string),
	}

	if label, ok := data["label"].(string); ok {
		job.Name = label
	}

	if command, ok := data["command"].(string); ok {
		job.Commands = append(job.Commands, command)
	} else if commands, ok := data["command"].([]interface{}); ok {
		for _, cmd := range commands {
			if cmdStr, ok := cmd.(string); ok {
				job.Commands = append(job.Commands, cmdStr)
			}
		}
	}

	return job
}

// ToJSON converts a pipeline to JSON
func (p *Pipeline) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML converts a pipeline to YAML
func (p *Pipeline) ToYAML() (string, error) {
	data, err := yaml.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
