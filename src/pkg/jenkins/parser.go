package jenkins

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// JobConfig represents a Jenkins job configuration
type JobConfig struct {
	XMLName        xml.Name       `xml:"project"`
	Description    string         `xml:"description,omitempty"`
	KeepDependencies bool         `xml:"keepDependencies"`
	Properties     []Property     `xml:"properties>hudson.model.ParametersDefinitionProperty>parameterDefinitions>*"`
	SCM            *SCMConfig     `xml:"scm"`
	Triggers       []Trigger      `xml:"triggers>*"`
	Builders       []Builder      `xml:"builders>*"`
	Publishers     []Publisher    `xml:"publishers>*"`
	BuildWrappers  []BuildWrapper `xml:"buildWrappers>*"`
	Disabled       bool           `xml:"disabled,omitempty"`
}

// Property represents a job property (parameters, etc.)
type Property struct {
	XMLName xml.Name
	Name    string `xml:"name"`
	Default string `xml:"defaultValue,omitempty"`
	Description string `xml:"description,omitempty"`
}

// SCMConfig represents source control configuration
type SCMConfig struct {
	XMLName xml.Name
	Class   string              `xml:"class,attr"`
	UserRemoteConfigs []UserRemoteConfig `xml:"userRemoteConfigs>hudson.plugins.git.UserRemoteConfig"`
	Branches []BranchSpec        `xml:"branches>hudson.plugins.git.BranchSpec"`
	DoGenerateSubmoduleConfigurations bool `xml:"doGenerateSubmoduleConfigurations"`
}

// UserRemoteConfig represents a Git remote configuration
type UserRemoteConfig struct {
	URL    string `xml:"url"`
	Name   string `xml:"name,omitempty"`
	Refspec string `xml:"refspec,omitempty"`
	CredentialsId string `xml:"credentialsId,omitempty"`
}

// BranchSpec represents a Git branch specification
type BranchSpec struct {
	Name string `xml:"name"`
}

// Trigger represents a build trigger
type Trigger struct {
	XMLName xml.Name
	Spec    string `xml:"spec,omitempty"`
}

// Builder represents a build step
type Builder struct {
	XMLName xml.Name
	Command string `xml:"command,omitempty"`
	Targets string `xml:"targets,omitempty"`
	POM     string `xml:"pom,omitempty"`
	Goals   string `xml:"goals,omitempty"`
	Script  string `xml:"script,omitempty"`
}

// Publisher represents a post-build action
type Publisher struct {
	XMLName xml.Name
	Pattern string `xml:"pattern,omitempty"`
	Artifacts string `xml:"artifacts,omitempty"`
}

// BuildWrapper represents a build wrapper
type BuildWrapper struct {
	XMLName xml.Name
}

// ParseJobConfig parses Jenkins XML job configuration
func ParseJobConfig(xmlData []byte) (*JobConfig, error) {
	var config JobConfig
	if err := xml.Unmarshal(xmlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse job config: %w", err)
	}
	return &config, nil
}

// ConvertToCASCI converts Jenkins job config to CASCI format
func (j *JobConfig) ConvertToCASCI() *CASCIProject {
	project := &CASCIProject{
		Name:        extractJobName(j.Description),
		Description: j.Description,
		Enabled:     !j.Disabled,
	}

	// Convert SCM configuration
	if j.SCM != nil && len(j.SCM.UserRemoteConfigs) > 0 {
		project.RepositoryURL = j.SCM.UserRemoteConfigs[0].URL

		if len(j.SCM.Branches) > 0 {
			branchName := j.SCM.Branches[0].Name
			// Clean up branch names like */main to just main
			branchName = strings.TrimPrefix(branchName, "*/")
			branchName = strings.TrimPrefix(branchName, "refs/heads/")
			project.Branch = branchName
		} else {
			project.Branch = "main"
		}
	}

	// Convert build steps
	project.BuildCommands = j.extractBuildCommands()

	// Convert triggers
	project.Triggers = j.extractTriggers()

	// Convert parameters
	project.Parameters = j.extractParameters()

	// Convert post-build actions
	project.Artifacts = j.extractArtifacts()

	return project
}

// extractBuildCommands extracts build commands from builders
func (j *JobConfig) extractBuildCommands() []string {
	var commands []string

	for _, builder := range j.Builders {
		switch builder.XMLName.Local {
		case "hudson.tasks.Shell":
			// Shell script
			if builder.Command != "" {
				commands = append(commands, builder.Command)
			}

		case "hudson.tasks.BatchFile":
			// Windows batch
			if builder.Command != "" {
				commands = append(commands, builder.Command)
			}

		case "hudson.tasks.Maven":
			// Maven build
			goals := builder.Goals
			if goals == "" {
				goals = "clean install"
			}
			pom := builder.POM
			if pom == "" {
				pom = "pom.xml"
			}
			commands = append(commands, fmt.Sprintf("mvn -f %s %s", pom, goals))

		case "hudson.tasks.Ant":
			// Ant build
			targets := builder.Targets
			if targets == "" {
				targets = "build"
			}
			commands = append(commands, fmt.Sprintf("ant %s", targets))

		case "hudson.plugins.gradle.Gradle":
			// Gradle build
			tasks := builder.Targets
			if tasks == "" {
				tasks = "build"
			}
			commands = append(commands, fmt.Sprintf("gradle %s", tasks))

		case "org.jenkinsci.plugins.docker.workflow.Docker":
			// Docker build
			if builder.Script != "" {
				commands = append(commands, builder.Script)
			}
		}
	}

	return commands
}

// extractTriggers extracts build triggers
func (j *JobConfig) extractTriggers() []string {
	var triggers []string

	for _, trigger := range j.Triggers {
		switch trigger.XMLName.Local {
		case "hudson.triggers.SCMTrigger":
			triggers = append(triggers, "scm")
		case "hudson.triggers.TimerTrigger":
			triggers = append(triggers, "schedule")
		case "com.cloudbees.jenkins.GitHubPushTrigger":
			triggers = append(triggers, "github_push")
		case "com.dabsquared.gitlabjenkins.GitLabPushTrigger":
			triggers = append(triggers, "gitlab_push")
		}
	}

	if len(triggers) == 0 {
		triggers = append(triggers, "manual")
	}

	return triggers
}

// extractParameters extracts build parameters
func (j *JobConfig) extractParameters() map[string]string {
	params := make(map[string]string)

	for _, prop := range j.Properties {
		if prop.Name != "" {
			params[prop.Name] = prop.Default
		}
	}

	return params
}

// extractArtifacts extracts artifact patterns
func (j *JobConfig) extractArtifacts() []string {
	var artifacts []string

	for _, publisher := range j.Publishers {
		switch publisher.XMLName.Local {
		case "hudson.tasks.ArtifactArchiver":
			if publisher.Artifacts != "" {
				// Split by comma
				patterns := strings.Split(publisher.Artifacts, ",")
				for _, pattern := range patterns {
					pattern = strings.TrimSpace(pattern)
					if pattern != "" {
						artifacts = append(artifacts, pattern)
					}
				}
			}
		}
	}

	return artifacts
}

// extractJobName extracts a reasonable job name from description
func extractJobName(description string) string {
	if description == "" {
		return "Imported Job"
	}

	// Take first line
	lines := strings.Split(description, "\n")
	name := strings.TrimSpace(lines[0])

	// Limit length
	if len(name) > 50 {
		name = name[:50]
	}

	if name == "" {
		name = "Imported Job"
	}

	return name
}

// CASCIProject represents a converted CASCI project
type CASCIProject struct {
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	RepositoryURL string            `json:"repository_url"`
	Branch        string            `json:"branch"`
	BuildCommands []string          `json:"build_commands"`
	Triggers      []string          `json:"triggers"`
	Parameters    map[string]string `json:"parameters,omitempty"`
	Artifacts     []string          `json:"artifacts,omitempty"`
	Enabled       bool              `json:"enabled"`
}

// Jenkinsfile represents a parsed Jenkinsfile (Groovy-based)
type Jenkinsfile struct {
	Pipeline     *PipelineBlock
	Agent        string
	Stages       []Stage
	Post         *PostBlock
	Environment  map[string]string
	Parameters   []Parameter
	Triggers     []string
	Tools        map[string]string
}

// PipelineBlock represents pipeline { } block
type PipelineBlock struct {
	Agent       string
	Stages      []Stage
	Environment map[string]string
}

// Stage represents a pipeline stage
type Stage struct {
	Name  string
	Steps []Step
	When  string
}

// Step represents a step within a stage
type Step struct {
	Type    string // sh, bat, script, etc.
	Content string
}

// PostBlock represents post { } actions
type PostBlock struct {
	Always  []Step
	Success []Step
	Failure []Step
	Cleanup []Step
}

// Parameter represents a build parameter
type Parameter struct {
	Name         string
	Type         string // string, choice, boolean, etc.
	DefaultValue string
	Description  string
	Choices      []string
}

// ParseJenkinsfile parses a Groovy Jenkinsfile
// Note: This is a simplified parser. Full Groovy parsing would require
// a complete Groovy parser, which is beyond the scope.
// This handles common declarative pipeline patterns.
func ParseJenkinsfile(content string) (*Jenkinsfile, error) {
	jf := &Jenkinsfile{
		Environment: make(map[string]string),
		Tools:       make(map[string]string),
	}

	// Simple line-by-line parsing for declarative pipelines
	lines := strings.Split(content, "\n")
	inPipeline := false
	inStage := false
	inSteps := false
	currentStage := &Stage{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "pipeline") {
			inPipeline = true
			continue
		}

		if !inPipeline {
			continue
		}

		// Parse agent
		if strings.HasPrefix(line, "agent") {
			jf.Agent = extractAgentValue(line)
		}

		// Parse stages
		if strings.HasPrefix(line, "stages") {
			continue
		}

		// Parse individual stage
		if strings.HasPrefix(line, "stage") {
			if inStage && currentStage.Name != "" {
				jf.Stages = append(jf.Stages, *currentStage)
			}
			inStage = true
			currentStage = &Stage{
				Name: extractStageName(line),
			}
		}

		// Parse steps
		if strings.HasPrefix(line, "steps") {
			inSteps = true
			continue
		}

		if inSteps && inStage {
			if strings.HasPrefix(line, "sh ") || strings.HasPrefix(line, "sh(") {
				step := Step{
					Type:    "sh",
					Content: extractShellCommand(line),
				}
				currentStage.Steps = append(currentStage.Steps, step)
			}
		}

		// End of stage
		if line == "}" && inStage {
			if inSteps {
				inSteps = false
			} else {
				jf.Stages = append(jf.Stages, *currentStage)
				currentStage = &Stage{}
				inStage = false
			}
		}
	}

	// Add last stage if exists
	if inStage && currentStage.Name != "" {
		jf.Stages = append(jf.Stages, *currentStage)
	}

	return jf, nil
}

// Helper functions for Jenkinsfile parsing
func extractAgentValue(line string) string {
	if strings.Contains(line, "any") {
		return "any"
	}
	if strings.Contains(line, "docker") {
		return "docker"
	}
	return "any"
}

func extractStageName(line string) string {
	// Extract stage name from: stage('Build') or stage("Build")
	start := strings.Index(line, "'")
	if start == -1 {
		start = strings.Index(line, "\"")
	}
	if start == -1 {
		return "Unnamed Stage"
	}

	end := strings.LastIndex(line, "'")
	if end == -1 {
		end = strings.LastIndex(line, "\"")
	}
	if end == -1 || end <= start {
		return "Unnamed Stage"
	}

	return line[start+1 : end]
}

func extractShellCommand(line string) string {
	// Extract command from: sh 'command' or sh "command" or sh(script: 'command')
	if strings.Contains(line, "script:") {
		start := strings.Index(line, "'")
		if start == -1 {
			start = strings.Index(line, "\"")
		}
		if start == -1 {
			return ""
		}

		end := strings.LastIndex(line, "'")
		if end == -1 {
			end = strings.LastIndex(line, "\"")
		}
		if end == -1 || end <= start {
			return ""
		}

		return line[start+1 : end]
	}

	// Simple sh 'command' format
	start := strings.Index(line, "'")
	if start == -1 {
		start = strings.Index(line, "\"")
	}
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(line, "'")
	if end == -1 {
		end = strings.LastIndex(line, "\"")
	}
	if end == -1 || end <= start {
		return ""
	}

	return line[start+1 : end]
}

// ConvertJenkinsfileToCASCI converts Jenkinsfile to CASCI pipeline
func (jf *Jenkinsfile) ConvertToCASCI() map[string]interface{} {
	pipeline := map[string]interface{}{
		"version": "1.0",
		"stages":  []map[string]interface{}{},
	}

	if len(jf.Environment) > 0 {
		pipeline["environment"] = map[string]interface{}{
			"variables": jf.Environment,
		}
	}

	// Convert stages
	for _, stage := range jf.Stages {
		casciStage := map[string]interface{}{
			"name": stage.Name,
			"jobs": []map[string]interface{}{
				{
					"name":     stage.Name,
					"commands": extractCommands(stage.Steps),
				},
			},
		}
		stages := pipeline["stages"].([]map[string]interface{})
		pipeline["stages"] = append(stages, casciStage)
	}

	return pipeline
}

func extractCommands(steps []Step) []string {
	var commands []string
	for _, step := range steps {
		if step.Content != "" {
			commands = append(commands, step.Content)
		}
	}
	return commands
}
