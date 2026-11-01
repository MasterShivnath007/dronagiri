package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// Pipx reads Python tools installed in isolated environments.
type Pipx struct{}

func (p *Pipx) ID() model.Source { return model.Pipx }

func (p *Pipx) Installed(r *runner.Runner) bool { return r.Available("pipx") }

type pipxDocument struct {
	Venvs map[string]struct {
		Metadata struct {
			VenvDir     string `json:"venv_dir"`
			MainPackage struct {
				PackageVersion string   `json:"package_version"`
				PackageOrURL   string   `json:"package_or_url"`
				Summary        string   `json:"summary"`
				Apps           []string `json:"apps"`
			} `json:"main_package"`
		} `json:"metadata"`
	} `json:"venvs"`
}

func (p *Pipx) Load(r *runner.Runner) ([]*model.Tool, error) {
	raw, err := r.Output("pipx", "list", "--json")
	if err != nil && strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var doc pipxDocument
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("could not read pipx metadata: %w", err)
	}

	home, _ := os.UserHomeDir()
	venvRoot := filepath.Join(home, ".local", "pipx", "venvs")
	if out, err := r.Output("pipx", "environment", "--value", "PIPX_LOCAL_VENVS"); err == nil {
		if trimmed := strings.TrimSpace(out); trimmed != "" {
			venvRoot = trimmed
		}
	}
	binDir := filepath.Join(home, ".local", "bin")

	names := make([]string, 0, len(doc.Venvs))
	for name := range doc.Venvs {
		names = append(names, name)
	}
	sort.Strings(names)

	var tools []*model.Tool
	var sizePaths []string

	for _, name := range names {
		entry := doc.Venvs[name]
		path := entry.Metadata.VenvDir
		if path == "" {
			path = filepath.Join(venvRoot, name)
		}

		tool := &model.Tool{
			ID:            string(model.Pipx) + ":" + name,
			Name:          name,
			Source:        model.Pipx,
			Version:       entry.Metadata.MainPackage.PackageVersion,
			Summary:       entry.Metadata.MainPackage.Summary,
			Path:          path,
			UserInstalled: true,
		}
		for _, app := range entry.Metadata.MainPackage.Apps {
			candidate := filepath.Join(binDir, app)
			if _, err := os.Lstat(candidate); err == nil {
				tool.Binaries = append(tool.Binaries, candidate)
			}
		}
		if stat, err := os.Stat(path); err == nil {
			tool.InstalledAt = stat.ModTime()
		}

		tools = append(tools, tool)
		sizePaths = append(sizePaths, path)
	}

	sizes := r.DuSizes(sizePaths)
	for _, tool := range tools {
		tool.SizeBytes = sizes[tool.Path]
	}
	return tools, nil
}

func (p *Pipx) UninstallCommand(t *model.Tool) (string, []string, error) {
	return "pipx", []string{"uninstall", t.Name}, nil
}

func (p *Pipx) UpgradeCommand(t *model.Tool) (string, []string, error) {
	return "pipx", []string{"upgrade", t.Name}, nil
}
