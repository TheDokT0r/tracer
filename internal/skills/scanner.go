package skills

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanSkills scans four locations under claudeDir for skill files and returns
// them sorted by source priority (user, command, project, plugin) then
// alphabetically by name within each group.
func ScanSkills(claudeDir string) ([]Skill, error) {
	var all []Skill

	user, err := scanUserSkills(claudeDir)
	if err != nil {
		return nil, err
	}
	all = append(all, user...)

	cmds, err := scanCommands(claudeDir)
	if err != nil {
		return nil, err
	}
	all = append(all, cmds...)

	proj, err := scanProjectSkills(claudeDir)
	if err != nil {
		return nil, err
	}
	all = append(all, proj...)

	plugins, err := scanPluginSkills(claudeDir)
	if err != nil {
		return nil, err
	}
	all = append(all, plugins...)

	sortSkills(all)
	return all, nil
}

// scanUserSkills scans {claudeDir}/skills/*/SKILL.md
func scanUserSkills(claudeDir string) ([]Skill, error) {
	skillsDir := filepath.Join(claudeDir, "skills")
	entries, err := os.ReadDir(skillsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		sk, err := readSkillFile(path, SourceUser, "", e.Name())
		if err != nil {
			continue // skip unreadable entries
		}
		result = append(result, sk)
	}
	return result, nil
}

// scanCommands scans {claudeDir}/commands/*.md
func scanCommands(claudeDir string) ([]Skill, error) {
	cmdDir := filepath.Join(claudeDir, "commands")
	entries, err := os.ReadDir(cmdDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []Skill
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(cmdDir, e.Name())
		baseName := strings.TrimSuffix(e.Name(), ".md")
		sk, err := readSkillFile(path, SourceCommand, "", baseName)
		if err != nil {
			continue
		}
		result = append(result, sk)
	}
	return result, nil
}

// scanProjectSkills decodes project directory names and scans
// {decoded_path}/.claude/commands/*.md for each project.
func scanProjectSkills(claudeDir string) ([]Skill, error) {
	projDir := filepath.Join(claudeDir, "projects")
	entries, err := os.ReadDir(projDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		decoded := decodeProjectPath(e.Name())
		cmdDir := filepath.Join(decoded, ".claude", "commands")
		cmdEntries, err := os.ReadDir(cmdDir)
		if err != nil {
			continue
		}
		for _, ce := range cmdEntries {
			if ce.IsDir() || !strings.HasSuffix(ce.Name(), ".md") {
				continue
			}
			path := filepath.Join(cmdDir, ce.Name())
			baseName := strings.TrimSuffix(ce.Name(), ".md")
			sk, err := readSkillFile(path, SourceProject, "", baseName)
			if err != nil {
				continue
			}
			result = append(result, sk)
		}
	}
	return result, nil
}

// scanPluginSkills scans {claudeDir}/plugins/cache/claude-plugins-official/*/
// picking the latest version directory (skipping those with .orphaned_at),
// then scanning skills/*/SKILL.md within it.
func scanPluginSkills(claudeDir string) ([]Skill, error) {
	pluginsBase := filepath.Join(claudeDir, "plugins", "cache", "claude-plugins-official")
	pluginEntries, err := os.ReadDir(pluginsBase)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []Skill
	for _, pe := range pluginEntries {
		if !pe.IsDir() {
			continue
		}
		pluginName := pe.Name()
		pluginDir := filepath.Join(pluginsBase, pluginName)

		versionDir, err := latestVersionDir(pluginDir)
		if err != nil || versionDir == "" {
			continue
		}

		skillsDir := filepath.Join(versionDir, "skills")
		skillEntries, err := os.ReadDir(skillsDir)
		if err != nil {
			continue
		}
		for _, se := range skillEntries {
			if !se.IsDir() {
				continue
			}
			path := filepath.Join(skillsDir, se.Name(), "SKILL.md")
			sk, err := readSkillFile(path, SourcePlugin, pluginName, se.Name())
			if err != nil {
				continue
			}
			sk.ReadOnly = true
			result = append(result, sk)
		}
	}
	return result, nil
}

// latestVersionDir returns the path of the latest version subdirectory inside
// pluginDir, skipping any directory that contains an .orphaned_at file.
// "Latest" is determined by lexicographic sort of directory names (semver).
func latestVersionDir(pluginDir string) (string, error) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(pluginDir, e.Name())
		if _, err := os.Stat(filepath.Join(dir, ".orphaned_at")); err == nil {
			continue // orphaned version
		}
		candidates = append(candidates, dir)
	}
	if len(candidates) == 0 {
		return "", nil
	}
	sort.Strings(candidates)
	return candidates[len(candidates)-1], nil
}

// decodeProjectPath converts a dash-encoded directory name back to a filesystem
// path. The encoding replaces '/' with '-' and strips the leading '/'.
// Example: "-Users-or-projects-Shapes" -> "/Users/or/projects/Shapes"
func decodeProjectPath(encoded string) string {
	return strings.ReplaceAll(encoded, "-", "/")
}

// readSkillFile reads a skill file and populates a Skill struct from its
// content and frontmatter.
func readSkillFile(path string, source Source, pluginName, fallbackName string) (Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Skill{}, err
	}

	name, description := parseFrontmatter(data)
	if name == "" {
		name = fallbackName
	}

	return Skill{
		Name:        name,
		Description: description,
		Source:      source,
		PluginName:  pluginName,
		Path:        path,
		Dir:         filepath.Dir(path),
		Size:        info.Size(),
		ReadOnly:    false,
	}, nil
}

// parseFrontmatter extracts name and description from YAML frontmatter
// delimited by --- lines. It uses simple string matching rather than a YAML
// library.
func parseFrontmatter(data []byte) (name, description string) {
	const delim = "---\n"

	// Must start with ---
	if !bytes.HasPrefix(data, []byte(delim)) {
		return "", ""
	}

	rest := data[len(delim):]
	end := bytes.Index(rest, []byte(delim))
	if end < 0 {
		return "", ""
	}

	block := string(rest[:end])
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return name, description
}

// sortSkills sorts by source priority (user < command < project < plugin),
// then alphabetically by name within each group.
func sortSkills(skills []Skill) {
	priority := map[Source]int{
		SourceUser:    0,
		SourceCommand: 1,
		SourceProject: 2,
		SourcePlugin:  3,
	}
	sort.SliceStable(skills, func(i, j int) bool {
		pi, pj := priority[skills[i].Source], priority[skills[j].Source]
		if pi != pj {
			return pi < pj
		}
		return strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
	})
}
