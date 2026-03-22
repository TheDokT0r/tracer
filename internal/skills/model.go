package skills

type Source string

const (
	SourceUser    Source = "user"
	SourceCommand Source = "command"
	SourceProject Source = "project"
	SourcePlugin  Source = "plugin"
)

type Skill struct {
	Name        string
	Description string
	Source      Source
	PluginName  string // only for plugin skills
	Path        string // full path to the skill file
	Dir         string // directory containing the skill
	Size        int64
	ReadOnly    bool // plugin skills are read-only
}
