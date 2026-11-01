package source

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// Cargo reads crates installed with `cargo install`.
type Cargo struct{}

func (c *Cargo) ID() model.Source { return model.Cargo }

func (c *Cargo) Installed(r *runner.Runner) bool { return r.Available("cargo") }

var (
	crateHeader = regexp.MustCompile(`^(\S+) v(\S+):\s*$`)
	crateBinary = regexp.MustCompile(`^\s+(\S+)\s*$`)
)

func (c *Cargo) Load(r *runner.Runner) ([]*model.Tool, error) {
	raw, err := r.Output("cargo", "install", "--list")
	if err != nil {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}

	binDir, _ := r.Output("cargo", "--version")
	_ = binDir
	home, _ := os.UserHomeDir()
	cargoBin := filepath.Join(home, ".cargo", "bin")

	var tools []*model.Tool
	var current *model.Tool
	var sizePaths []string

	flush := func() {
		if current == nil {
			return
		}
		tools = append(tools, current)
		if current.Path != "" {
			sizePaths = append(sizePaths, current.Path)
		}
		current = nil
	}

	for _, line := range strings.Split(raw, "\n") {
		if header := crateHeader.FindStringSubmatch(line); header != nil {
			flush()
			current = &model.Tool{
				ID:            string(model.Cargo) + ":" + header[1],
				Name:          header[1],
				Source:        model.Cargo,
				Version:       header[2],
				UserInstalled: true,
			}
			continue
		}
		if current == nil {
			continue
		}
		if binary := crateBinary.FindStringSubmatch(line); binary != nil {
			path := filepath.Join(cargoBin, binary[1])
			current.Binaries = append(current.Binaries, path)
			if current.Path == "" {
				current.Path = path
			}
			if current.InstalledAt.IsZero() {
				if stat, err := os.Stat(path); err == nil {
					current.InstalledAt = stat.ModTime()
					current.SizeBytes += stat.Size()
				}
			}
		}
	}
	flush()

	sizes := r.DuSizes(sizePaths)
	for _, tool := range tools {
		if size := sizes[tool.Path]; size > 0 {
			tool.SizeBytes = size
		}
	}
	return tools, nil
}

func (c *Cargo) UninstallCommand(t *model.Tool) (string, []string, error) {
	return "cargo", []string{"uninstall", t.Name}, nil
}

func (c *Cargo) UpgradeCommand(t *model.Tool) (string, []string, error) {
	return "cargo", []string{"install", "--force", t.Name}, nil
}
