package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// Brew reads Homebrew formulae and casks.
type Brew struct{}

func (b *Brew) ID() model.Source { return model.Brew }

func (b *Brew) Installed(r *runner.Runner) bool { return r.Available("brew") }

type brewDocument struct {
	Formulae []struct {
		Name         string   `json:"name"`
		Desc         string   `json:"desc"`
		Homepage     string   `json:"homepage"`
		Dependencies []string `json:"dependencies"`
		Outdated     bool     `json:"outdated"`
		Versions     struct {
			Stable string `json:"stable"`
		} `json:"versions"`
		Installed []struct {
			Version            string `json:"version"`
			Time               int64  `json:"time"`
			InstalledOnRequest bool   `json:"installed_on_request"`
		} `json:"installed"`
	} `json:"formulae"`
	Casks []struct {
		Token         string           `json:"token"`
		Name          []string         `json:"name"`
		Desc          string           `json:"desc"`
		Homepage      string           `json:"homepage"`
		Version       string           `json:"version"`
		Installed     string           `json:"installed"`
		Outdated      bool             `json:"outdated"`
		InstalledTime int64            `json:"installed_time"`
		Artifacts     []map[string]any `json:"artifacts"`
	} `json:"casks"`
}

func (b *Brew) Load(r *runner.Runner) ([]*model.Tool, error) {
	raw, err := r.Output("brew", "info", "--json=v2", "--installed")
	if err != nil {
		return nil, fmt.Errorf("%s", firstLine(err.Error()))
	}

	var doc brewDocument
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("could not read Homebrew metadata: %w", err)
	}

	prefix := "/opt/homebrew"
	if out, err := r.Output("brew", "--prefix"); err == nil {
		if trimmed := strings.TrimSpace(out); trimmed != "" {
			prefix = trimmed
		}
	}

	// Homebrew's own JSON lists *declared* dependencies. What a package
	// actually links against at runtime can differ, so ask brew directly for
	// every formula the user installed. That is the difference between
	// "aom is unused" and "aom is needed by mpv".
	var roots []string
	for _, f := range doc.Formulae {
		if len(f.Installed) == 0 {
			continue
		}
		if f.Installed[len(f.Installed)-1].InstalledOnRequest {
			roots = append(roots, f.Name)
		}
	}
	runtimeDeps := b.runtimeDependencies(r, roots)

	var tools []*model.Tool
	candidates := make(map[string][]string)

	for _, f := range doc.Formulae {
		if len(f.Installed) == 0 {
			continue
		}
		current := f.Installed[len(f.Installed)-1]

		tool := &model.Tool{
			ID:            string(model.Brew) + ":" + f.Name,
			Name:          f.Name,
			Source:        model.Brew,
			Version:       current.Version,
			Summary:       f.Desc,
			Homepage:      f.Homepage,
			Path:          filepath.Join(prefix, "Cellar", f.Name, current.Version),
			InstalledAt:   epoch(current.Time),
			UserInstalled: current.InstalledOnRequest,
		}
		if f.Outdated && f.Versions.Stable != "" && f.Versions.Stable != current.Version {
			tool.Latest = f.Versions.Stable
		}

		deps := f.Dependencies
		if resolved, ok := runtimeDeps[f.Name]; ok && len(resolved) > 0 {
			deps = resolved
		}
		for _, dep := range deps {
			if dep == f.Name {
				continue
			}
			tool.Dependencies = append(tool.Dependencies, string(model.Brew)+":"+dep)
		}

		for _, sub := range []string{"bin", "sbin"} {
			linked := filepath.Join(prefix, sub, f.Name)
			if _, err := os.Lstat(linked); err == nil {
				tool.Binaries = append(tool.Binaries, linked)
			}
		}

		tools = append(tools, tool)
		candidates[tool.ID] = []string{tool.Path}
	}

	for _, c := range doc.Casks {
		version := c.Installed
		if version == "" {
			version = c.Version
		}
		if version == "" {
			continue
		}

		tool := &model.Tool{
			ID:            string(model.Cask) + ":" + c.Token,
			Name:          c.Token,
			Source:        model.Cask,
			Version:       version,
			Summary:       c.Desc,
			Homepage:      c.Homepage,
			InstalledAt:   epoch(c.InstalledTime),
			UserInstalled: true,
			Protected:     c.Token == "homebrew-app",
		}
		if tool.Protected {
			tool.ProtectReason = "this app is how you manage Homebrew"
		}
		if c.Outdated && c.Version != "" && c.Version != version {
			tool.Latest = c.Version
		}

		appPath, binaries, targets := caskArtifacts(c.Artifacts)
		tool.Path = appPath
		tool.Binaries = binaries

		paths := append([]string{}, targets...)
		paths = append(paths, binaries...)
		if appPath != "" {
			paths = append(paths, appPath)
		}
		// Caskroom keeps the staged copy; for command-line casks it is often
		// the only place that reflects real disk use.
		paths = append(paths, filepath.Join(prefix, "Caskroom", c.Token))
		candidates[tool.ID] = dedupe(paths)

		tools = append(tools, tool)
	}

	var allPaths []string
	for _, paths := range candidates {
		allPaths = append(allPaths, paths...)
	}
	sizes := r.DuSizes(dedupe(allPaths))
	for _, tool := range tools {
		for _, path := range candidates[tool.ID] {
			if size := sizes[path]; size > tool.SizeBytes {
				tool.SizeBytes = size
			}
		}
	}

	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools, nil
}

// runtimeDependencies asks brew for the real dependency closure of each
// formula, a handful at a time. Each call is fast but there is one per root.
func (b *Brew) runtimeDependencies(r *runner.Runner, formulae []string) map[string][]string {
	result := make(map[string][]string, len(formulae))
	if len(formulae) == 0 {
		return result
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	limit := make(chan struct{}, 8)

	for _, name := range formulae {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			out, err := r.Output("brew", "deps", name)
			if err != nil {
				return
			}

			var deps []string
			for _, line := range strings.Split(out, "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					deps = append(deps, line)
				}
			}

			mu.Lock()
			result[name] = deps
			mu.Unlock()
		}(name)
	}
	wg.Wait()
	return result
}

func (b *Brew) UninstallCommand(t *model.Tool) (string, []string, error) {
	if t.Source == model.Cask {
		return "brew", []string{"uninstall", "--cask", t.Name}, nil
	}
	return "brew", []string{"uninstall", t.Name}, nil
}

func (b *Brew) UpgradeCommand(t *model.Tool) (string, []string, error) {
	if t.Source == model.Cask {
		return "brew", []string{"upgrade", "--cask", t.Name}, nil
	}
	return "brew", []string{"upgrade", t.Name}, nil
}

// caskArtifacts works out where a cask put its files.
func caskArtifacts(artifacts []map[string]any) (appPath string, binaries, targets []string) {
	for _, artifact := range artifacts {
		if apps, ok := artifact["app"].([]any); ok && len(apps) > 0 {
			if name, ok := apps[0].(string); ok && appPath == "" {
				appPath = "/Applications/" + name
			}
		}
		if bins, ok := artifact["binary"].([]any); ok {
			for _, entry := range bins {
				if name, ok := entry.(string); ok {
					binaries = append(binaries, name)
				}
			}
		}
		if target, ok := artifact["target"].(string); ok && target != "" {
			targets = append(targets, target)
		}
	}
	return appPath, binaries, targets
}

func dedupe(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func epoch(seconds int64) time.Time {
	if seconds <= 0 {
		return time.Time{}
	}
	return time.Unix(seconds, 0)
}
