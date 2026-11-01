// Package source knows how to read one package ecosystem.
package source

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
)

// Loader reads the packages installed by one manager.
type Loader interface {
	// ID is the manager this loader speaks for.
	ID() model.Source
	// Load returns every installed package.
	Load(r *runner.Runner) ([]*model.Tool, error)
	// Installed reports whether the manager itself is present.
	Installed(r *runner.Runner) bool
	// UninstallCommand returns the program and arguments that remove one tool.
	UninstallCommand(t *model.Tool) (string, []string, error)
	// UpgradeCommand returns the program and arguments that update one tool.
	UpgradeCommand(t *model.Tool) (string, []string, error)
}

var (
	// ErrProtected is returned when a package must not be removed.
	ErrProtected = fmt.Errorf("this package is part of macOS and cannot be removed")
	// ErrUnsupported is returned when a manager has no safe removal path.
	ErrUnsupported = fmt.Errorf("this manager does not support removal from dronagiri")
)

// all returns the loaders in display order.
func all() []Loader {
	return []Loader{
		&Brew{},
		&NPM{},
		&Gem{},
		&Pipx{},
		&Cargo{},
		&GoBin{},
	}
}

// Progress is called as each manager finishes.
type Progress func(source model.Source, tools int, err error)

// LoadAll reads every available manager concurrently and assembles the
// inventory. A failing manager never hides the others.
func LoadAll(r *runner.Runner, progress Progress) *model.Inventory {
	inv := &model.Inventory{LoadedAt: time.Now()}

	loaders := all()
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, loader := range loaders {
		if !loader.Installed(r) {
			continue
		}
		wg.Add(1)
		go func(l Loader) {
			defer wg.Done()
			tools, err := l.Load(r)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				inv.Errors = append(inv.Errors, model.SourceError{Source: l.ID(), Err: err})
			} else {
				inv.Tools = append(inv.Tools, tools...)
			}
			if progress != nil {
				progress(l.ID(), len(tools), err)
			}
		}(loader)
	}

	wg.Wait()
	inv.Finish()
	enrichUsage(r, inv.Tools)
	markBroken(inv)
	return inv
}

// markBroken flags packages whose recorded location no longer exists. This is
// the residue left behind by a manual delete or an interrupted uninstall.
func markBroken(inv *model.Inventory) {
	for _, tool := range inv.Tools {
		if tool.Path == "" {
			continue
		}
		if _, err := os.Stat(tool.Path); err != nil {
			tool.Broken = true
		}
	}
}

// FindLoader returns the loader responsible for a tool.
func FindLoader(t *model.Tool) Loader {
	for _, loader := range all() {
		if loader.ID() == t.Source {
			return loader
		}
	}
	return nil
}

// Uninstall removes one tool using its own manager.
func Uninstall(r *runner.Runner, t *model.Tool) (string, error) {
	if t.Protected {
		return "", fmt.Errorf("%s: %w", t.Name, ErrProtected)
	}
	loader := FindLoader(t)
	if loader == nil {
		return "", ErrUnsupported
	}
	program, args, err := loader.UninstallCommand(t)
	if err != nil {
		return "", err
	}
	return r.Stream(program, args...)
}

// Upgrade updates one tool using its own manager.
func Upgrade(r *runner.Runner, t *model.Tool) (string, error) {
	loader := FindLoader(t)
	if loader == nil {
		return "", ErrUnsupported
	}
	program, args, err := loader.UpgradeCommand(t)
	if err != nil {
		return "", err
	}
	return r.Stream(program, args...)
}

// firstLine is a small helper for adapters that report errors.
func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
