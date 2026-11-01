package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// GoBin reads binaries installed with `go install`. They carry their module
// path in the executable metadata, which is how we learn their version.
type GoBin struct{}

func (g *GoBin) ID() model.Source { return model.GoBin }

func (g *GoBin) Installed(r *runner.Runner) bool { return r.Available("go") }

func (g *GoBin) Load(r *runner.Runner) ([]*model.Tool, error) {
	gopathOut, err := r.Output("go", "env", "GOPATH")
	if err != nil {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}
	gopath := strings.TrimSpace(gopathOut)
	if gopath == "" {
		return nil, nil
	}

	binDir := filepath.Join(gopath, "bin")
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return nil, nil // no Go binaries yet is not an error
	}

	var tools []*model.Tool
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(binDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0o111 == 0 {
			continue
		}

		tool := &model.Tool{
			ID:            string(model.GoBin) + ":" + entry.Name(),
			Name:          entry.Name(),
			Source:        model.GoBin,
			Path:          path,
			Binaries:      []string{path},
			InstalledAt:   info.ModTime(),
			SizeBytes:     info.Size(),
			UserInstalled: true,
		}

		if out, err := r.Output("go", "version", "-m", path); err == nil {
			for _, line := range strings.Split(out, "\n") {
				fields := strings.Fields(line)
				if len(fields) < 3 {
					continue
				}
				if fields[0] != "mod" && fields[0] != "path" {
					continue
				}
				tool.Summary = "Go module " + fields[1]
				if len(fields) > 2 && strings.HasPrefix(fields[2], "v") {
					tool.Version = fields[2]
				}
				if tool.Version == "" {
					tool.Version = "(devel)"
				}
				break
			}
		}
		if tool.Version == "" {
			tool.Version = "unknown"
		}

		tools = append(tools, tool)
	}

	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools, nil
}

func (g *GoBin) UninstallCommand(t *model.Tool) (string, []string, error) {
	// Go binaries are plain files with no package database, so removal is
	// simply deleting the executable.
	return "rm", []string{"-f", t.Path}, nil
}

func (g *GoBin) UpgradeCommand(t *model.Tool) (string, []string, error) {
	return "", nil, fmt.Errorf("reinstall with `go install <module>@latest`")
}
