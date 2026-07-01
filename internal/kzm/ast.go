package kzm

// Program represents the root AST node of a parsed KZM configuration file.
type Program struct {
	Meta         *Meta
	Triggers     []*Trigger
	Secrets      *Secrets
	Variables    []*Variable
	Environments []*Environment
	Before       *Phase
	Deploy       *DeployTarget
	After        *Phase
	Notifiers    []*Notifier
	Plugins      []string
	KeepLast     int
	Line         int
}

// Meta holds descriptive information about the project.
type Meta struct {
	Name        string
	Version     string
	Description string
}

// Trigger defines what events trigger the pipeline (e.g. branch push, schedule, etc.).
type Trigger struct {
	Source string // github, gitlab, folder, cron, manual
	Repo   string
	Branch string
	Path   string
	Cron   string
	Name   string
	Line   int
}

// Secrets contains the path to the encrypted secrets store used by the program.
type Secrets struct {
	Path string
	Line int
}

// Variable represents a static key-value variable pair.
type Variable struct {
	Key   string
	Value string
	Line  int
}

// Phase contains a sequence of steps executed as a distinct block of the pipeline.
type Phase struct {
	Steps []*Step
	Line  int
}

// Step represents an individual command to execute, along with validation checks and conditions.
type Step struct {
	Command    string
	MustPass   bool
	OrRollback bool
	Condition  *Condition
	Line       int
}

// Condition holds environment variables to match before a step execution.
type Condition struct {
	EnvName  string
	EnvValue string
}

// DeployTarget defines the deployment type, target identifier, and configuration parameters.
type DeployTarget struct {
	Type     string            // ssh, lambda, cpanel, s3, docker, vps, local, ftp, vercel
	Name     string            // name or host of target
	Options  map[string]string // configuration options (e.g. key, restart, region, etc.)
	EnvVars  map[string]string // env block variables
	Line     int
}

// Notifier stores notification channels (Slack, Email) and their respective endpoints/addresses.
type Notifier struct {
	Channel string // slack, email
	Target  string // channel string or email address
	Line    int
}

// Environment represents environment-specific override blocks in the config.
type Environment struct {
	Name   string
	Target *DeployTarget
	Line   int
}
