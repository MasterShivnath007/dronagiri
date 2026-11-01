// Package model holds the data types shared by every package source.
package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Source identifies the tool that installed a package.
type Source string

const (
	Brew  Source = "brew"
	Cask  Source = "cask"
	NPM   Source = "npm"
	Gem   Source = "gem"
	Cargo Source = "cargo"
	Pipx  Source = "pipx"
	GoBin Source = "go"
)

// Label is the short word shown in the interface.
func (s Source) Label() string {
	switch s {
	case Brew:
		return "brew"
	case Cask:
		return "cask"
	case NPM:
		return "npm"
	case Gem:
		return "gem"
	case Cargo:
		return "cargo"
	case Pipx:
		return "pipx"
	case GoBin:
		return "go"
	}
	return string(s)
}

// Tool is one installed package.
type Tool struct {
	ID            string
	Name          string
	Source        Source
	Version       string
	Latest        string
	Summary       string
	Homepage      string
	Path          string
	Binaries      []string
	InstalledAt   time.Time
	SizeBytes     int64
	UserInstalled bool
	Protected     bool
	ProtectReason string
	Dependencies  []string
	// Broken means the manager still lists the package but its files are
	// gone — usually a manual delete or a half-finished uninstall.
	Broken bool
	// LastUsed is the best evidence we have of the last run. macOS does not
	// update access times, so this comes from Spotlight (applications) or from
	// shell history (commands). Zero means no evidence was found.
	LastUsed       time.Time
	LastUsedSeen   bool
	LastUsedSource string
}

// DisplayName is the name plus the source, used when a package name alone is
// ambiguous (for example "json" exists in both brew and gem).
func (t *Tool) DisplayName() string {
	return t.Name
}

// Outdated reports whether a newer version is known to exist.
func (t *Tool) Outdated() bool {
	return t.Latest != "" && t.Latest != t.Version
}

// IsDependency reports whether this package was pulled in to satisfy another.
func (t *Tool) IsDependency() bool {
	return !t.UserInstalled
}

// SourceError records a package manager that could not be read.
type SourceError struct {
	Source Source
	Err    error
}

func (e SourceError) Error() string {
	return fmt.Sprintf("%s: %v", e.Source.Label(), e.Err)
}

// Inventory is the complete picture of the machine.
type Inventory struct {
	Tools      []*Tool
	RequiredBy map[string][]string
	Errors     []SourceError
	LoadedAt   time.Time
	TotalSize  int64
}

// Finish computes reverse dependencies and totals. It must be called once all
// tools have been appended.
func (inv *Inventory) Finish() {
	inv.RequiredBy = make(map[string][]string)
	inv.TotalSize = 0

	byID := make(map[string]*Tool, len(inv.Tools))
	for _, tool := range inv.Tools {
		byID[tool.ID] = tool
		inv.TotalSize += tool.SizeBytes
	}

	// For every package the user asked for, walk its dependency tree and
	// record that the dependency is required by that package.
	for _, root := range inv.Tools {
		if !root.UserInstalled {
			continue
		}
		seen := make(map[string]bool)
		stack := append([]string(nil), root.Dependencies...)
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] || id == root.ID {
				continue
			}
			seen[id] = true
			inv.RequiredBy[id] = appendUnique(inv.RequiredBy[id], root.Name)
			if dep, ok := byID[id]; ok {
				stack = append(stack, dep.Dependencies...)
			}
		}
	}

	for id := range inv.RequiredBy {
		sort.Strings(inv.RequiredBy[id])
	}
	sort.Slice(inv.Tools, func(i, j int) bool {
		if inv.Tools[i].UserInstalled != inv.Tools[j].UserInstalled {
			return inv.Tools[i].UserInstalled
		}
		return strings.ToLower(inv.Tools[i].Name) < strings.ToLower(inv.Tools[j].Name)
	})
}

func appendUnique(list []string, value string) []string {
	for _, item := range list {
		if item == value {
			return list
		}
	}
	return append(list, value)
}

// HumanBytes renders a byte count the way a person would say it.
func HumanBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < 4; n /= unit {
		div *= unit
		exp++
	}
	value := float64(bytes) / float64(div)
	if value >= 100 {
		return fmt.Sprintf("%.0f %cB", value, "KMGTP"[exp])
	}
	return fmt.Sprintf("%.1f %cB", value, "KMGTP"[exp])
}

// HumanDate renders an install date without the time of day.
func HumanDate(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("02 Jan 2006")
}

// HumanDateShort is the compact form used in table columns.
func HumanDateShort(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("02 Jan 06")
}

// CompareVersions orders dot-separated versions numerically, so 1.10 comes
// after 1.9 rather than before it.
func CompareVersions(a, b string) int {
	as := strings.FieldsFunc(a, func(r rune) bool { return r == '.' || r == '-' || r == '_' })
	bs := strings.FieldsFunc(b, func(r rune) bool { return r == '.' || r == '-' || r == '_' })

	for i := 0; i < len(as) || i < len(bs); i++ {
		if i >= len(as) {
			return -1
		}
		if i >= len(bs) {
			return 1
		}

		ai, aErr := strconv.Atoi(as[i])
		bi, bErr := strconv.Atoi(bs[i])
		switch {
		case aErr == nil && bErr == nil && ai != bi:
			if ai > bi {
				return 1
			}
			return -1
		case aErr == nil && bErr != nil:
			return 1
		case aErr != nil && bErr == nil:
			return -1
		case as[i] != bs[i]:
			if as[i] > bs[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
