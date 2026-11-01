package source

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// Gem reads RubyGems. Gems shipped with macOS are skipped: they are part of
// the operating system, not something the user chose to install.
type Gem struct{}

func (g *Gem) ID() model.Source { return model.Gem }

func (g *Gem) Installed(r *runner.Runner) bool { return r.Available("gem") }

var (
	gemListLine   = regexp.MustCompile(`^(\S+) \(([^)]+)\)\s*$`)
	gemSummary    = regexp.MustCompile(`\.summary\s*=\s*"([^"]*)"`)
	gemDependency = regexp.MustCompile(`add(?:_runtime)?_dependency\s*\(?\s*["']([A-Za-z0-9_.\-]+)["']`)
)

func (g *Gem) Load(r *runner.Runner) ([]*model.Tool, error) {
	homeOut, err := r.Output("gem", "env", "home")
	if err != nil {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}
	home := strings.TrimSpace(homeOut)
	if home == "" {
		return nil, nil
	}

	raw, err := r.Output("gem", "list", "--local")
	if err != nil {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}

	systemGems := strings.HasPrefix(home, "/Library/Ruby") ||
		strings.HasPrefix(home, "/System") ||
		strings.HasPrefix(home, "/usr/lib/ruby")

	var tools []*model.Tool
	var sizePaths []string

	for _, line := range strings.Split(raw, "\n") {
		match := gemListLine.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}
		name := match[1]

		version, onlyDefault := pickGemVersion(match[2])
		if version == "" || onlyDefault {
			continue // a gem that ships with Ruby itself
		}

		// Gems inside macOS' own Ruby are part of the operating system, and
		// some are listed without any files on disk. Neither belongs in a
		// list of things the user installed.
		if systemGems {
			continue
		}

		specPath := filepath.Join(home, "specifications", name+"-"+version+".gemspec")
		gemPath := filepath.Join(home, "gems", name+"-"+version)
		if _, err := os.Stat(gemPath); err != nil {
			continue
		}

		tool := &model.Tool{
			ID:            string(model.Gem) + ":" + name,
			Name:          name,
			Source:        model.Gem,
			Version:       version,
			Path:          gemPath,
			UserInstalled: true,
		}

		if data, err := os.ReadFile(specPath); err == nil {
			text := string(data)
			if m := gemSummary.FindStringSubmatch(text); m != nil {
				tool.Summary = m[1]
			}
			for _, m := range gemDependency.FindAllStringSubmatch(text, -1) {
				if m[1] != name {
					tool.Dependencies = append(tool.Dependencies, string(model.Gem)+":"+m[1])
				}
			}
			sort.Strings(tool.Dependencies)
			if stat, err := os.Stat(specPath); err == nil {
				tool.InstalledAt = stat.ModTime()
			}
		}

		binary := filepath.Join(filepath.Dir(filepath.Dir(home)), "bin", name)
		if _, err := os.Lstat(binary); err == nil {
			tool.Binaries = append(tool.Binaries, binary)
		}

		tools = append(tools, tool)
		sizePaths = append(sizePaths, gemPath)
	}

	sizes := r.DuSizes(sizePaths)
	for _, tool := range tools {
		tool.SizeBytes = sizes[tool.Path]
	}
	return tools, nil
}

// pickGemVersion chooses the newest non-default version. It reports whether
// every listed version is a Ruby default.
func pickGemVersion(list string) (string, bool) {
	var versions []string
	defaults := 0
	for _, part := range strings.Split(list, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "default:") {
			defaults++
			continue
		}
		versions = append(versions, part)
	}
	if len(versions) == 0 {
		return "", defaults > 0
	}
	sort.Slice(versions, func(i, j int) bool { return model.CompareVersions(versions[i], versions[j]) > 0 })
	return versions[0], false
}

func (g *Gem) UninstallCommand(t *model.Tool) (string, []string, error) {
	return "gem", []string{"uninstall", t.Name, "-v", t.Version, "-x"}, nil
}

func (g *Gem) UpgradeCommand(t *model.Tool) (string, []string, error) {
	return "gem", []string{"update", t.Name}, nil
}
