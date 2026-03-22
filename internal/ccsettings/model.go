package ccsettings

// Scope identifies where a settings file lives.
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
	ScopeLocal   Scope = "local"
)

// SettingsFile represents one settings.json with its permissions.
type SettingsFile struct {
	Scope       Scope
	Path        string
	ProjectPath string // empty for global
	Permissions Permissions
}

// Permissions holds the allow/deny/ask lists.
type Permissions struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

// PermRule is a single permission rule for display.
type PermRule struct {
	Rule   string
	List   string // "allow" or "deny"
	Source *SettingsFile
}
