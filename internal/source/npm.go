package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

var unicodeEscape = regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)

// NPM reads globally installed Node packages.
type NPM struct{}

func (n *NPM) ID() model.Source { return model.NPM }

func (n *NPM) Installed(r *runner.Runner) bool { return r.Available("npm") }

type npmManifest struct {
	Description  string            `json:"description"`
	Dependencies map[string]string `json:"dependencies"`
	Bin          any               `json:"bin"`
}

func (n *NPM) Load(r *runner.Runner) ([]*model.Tool, error) {
	rootOut, err := r.Output("npm", "root", "-g")
	if err != nil {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}
	root := strings.TrimSpace(rootOut)
	if root == "" {
		return nil, nil
	}

	raw, err := r.Output("npm", "ls", "-g", "--depth=0", "--json")
	if err != nil && strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}

	var doc struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("could not read npm metadata: %w", err)
	}

	names := make([]string, 0, len(doc.Dependencies))
	for name := range doc.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	binDir := filepath.Join(filepath.Dir(root), "bin")
	var tools []*model.Tool
	var sizePaths []string

	for _, name := range names {
		info := doc.Dependencies[name]
		pkgDir := filepath.Join(root, name)

		tool := &model.Tool{
			ID:            string(model.NPM) + ":" + name,
			Name:          name,
			Source:        model.NPM,
			Version:       info.Version,
			Path:          pkgDir,
			UserInstalled: true,
		}

		if data, err := os.ReadFile(filepath.Join(pkgDir, "package.json")); err == nil {
			var manifest npmManifest
			if json.Unmarshal(data, &manifest) == nil {
				tool.Summary = decodeEscapes(manifest.Description)
				for dep := range manifest.Dependencies {
					if _, ok := doc.Dependencies[dep]; ok {
						tool.Dependencies = append(tool.Dependencies, string(model.NPM)+":"+dep)
					}
				}
				sort.Strings(tool.Dependencies)
				for _, bin := range npmBinNames(manifest.Bin, name) {
					candidate := filepath.Join(binDir, bin)
					if _, err := os.Lstat(candidate); err == nil {
						tool.Binaries = append(tool.Binaries, candidate)
					}
				}
			}
		}

		if stat, err := os.Stat(pkgDir); err == nil {
			tool.InstalledAt = stat.ModTime()
		}

		tools = append(tools, tool)
		sizePaths = append(sizePaths, pkgDir)
	}

	sizes := r.DuSizes(sizePaths)
	for _, tool := range tools {
		tool.SizeBytes = sizes[tool.Path]
	}
	return tools, nil
}

func npmBinNames(bin any, fallback string) []string {
	switch value := bin.(type) {
	case string:
		return []string{value}
	case map[string]any:
		names := make([]string, 0, len(value))
		for key := range value {
			names = append(names, key)
		}
		return names
	}
	return []string{fallback}
}

// decodeEscapes turns literal `\u2014` sequences found in some package.json
// descriptions into real characters.
func decodeEscapes(s string) string {
	return unicodeEscape.ReplaceAllStringFunc(s, func(match string) string {
		value, err := strconv.ParseInt(match[2:], 16, 32)
		if err != nil {
			return match
		}
		return string(rune(value))
	})
}

func (n *NPM) UninstallCommand(t *model.Tool) (string, []string, error) {
	return "npm", []string{"uninstall", "-g", t.Name}, nil
}

func (n *NPM) UpgradeCommand(t *model.Tool) (string, []string, error) {
	return "npm", []string{"install", "-g", t.Name + "@latest"}, nil
}
