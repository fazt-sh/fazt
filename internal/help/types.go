package help

// CommandDoc represents the parsed structure of a CLI help markdown file
type CommandDoc struct {
	// Frontmatter fields
	Command     string     `yaml:"command"`
	Description string     `yaml:"description"`
	Syntax      string     `yaml:"syntax"`
	Version     string     `yaml:"version"`
	Updated     string     `yaml:"updated"`
	Category    string     `yaml:"category"`
	Arguments   []Argument `yaml:"arguments"`
	Flags       []Flag     `yaml:"flags"`
	Examples    []Example  `yaml:"examples"`
	Related     []Related  `yaml:"related"`
	Errors      []Error    `yaml:"errors"`
	Peer        *PeerInfo  `yaml:"peer"`

	// Markdown body (extended documentation)
	Body string `yaml:"-"`
}

// Argument represents a command argument
type Argument struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
}

// Flag represents a command flag
type Flag struct {
	Name        string `yaml:"name"`
	Short       string `yaml:"short"`
	Type        string `yaml:"type"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
}

// Example represents a usage example
type Example struct {
	Title       string `yaml:"title"`
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// Related represents a related command
type Related struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// Error represents a common error and its solution
type Error struct {
	Code     string `yaml:"code"`
	Message  string `yaml:"message"`
	Solution string `yaml:"solution"`
}

// PeerInfo describes peer support for a command
type PeerInfo struct {
	Supported    bool          `yaml:"supported"`
	Local        bool          `yaml:"local"`
	Remote       bool          `yaml:"remote"`
	Syntax       string        `yaml:"syntax"`
	RemovedFlags []RemovedFlag `yaml:"removed_flags"`
}

// RemovedFlag documents a flag that was removed
type RemovedFlag struct {
	Name        string `yaml:"name"`
	RemovedIn   string `yaml:"removed_in"`
	Replacement string `yaml:"replacement"`
}
