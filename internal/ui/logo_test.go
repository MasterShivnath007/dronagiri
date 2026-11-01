package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/MasterShivnath007/dronagiri/internal/model"
)

func TestLogoPicksTheRightSize(t *testing.T) {
	if got := len(logoLines(120)); got != 16 {
		t.Errorf("wide terminal should get the 16-line emblem, got %d", got)
	}
	if got := len(logoLines(70)); got != 12 {
		t.Errorf("narrow terminal should get the 12-line emblem, got %d", got)
	}
	if got := len(logoLines(80)); got != 12 {
		t.Errorf("80 columns cannot hold the wide emblem, got %d lines", got)
	}
	if got := len(logoLines(50)); got != 0 {
		t.Errorf("tiny terminal should get no emblem, got %d lines", got)
	}
}

func TestLogoFitsItsTerminalClass(t *testing.T) {
	if width := logoWidth(120); width > 84 {
		t.Errorf("wide emblem is %d columns, too wide for the windows it is used in", width)
	}
	if width := logoWidth(70); width > 60 {
		t.Errorf("narrow emblem is %d columns, does not fit 60", width)
	}
}

func TestHeaderIsOneLine(t *testing.T) {
	m := &Model{width: 120, height: 40}
	if strings.Contains(m.header(), "\n") {
		t.Error("browsing header must stay on one line")
	}
}

func TestSplashShowsTheEmblemAndName(t *testing.T) {
	m := &Model{width: 120, height: 42}
	splash := m.splashView()

	if !strings.Contains(splash, "Dronagiri") {
		t.Error("splash is missing the name")
	}
	if !strings.Contains(splash, slogan) {
		t.Error("splash is missing the slogan")
	}
	if lines := strings.Count(splash, "\n"); lines > m.height {
		t.Errorf("splash is %d lines tall but the terminal is %d", lines, m.height)
	}
}

func TestSloganStartsCapitalised(t *testing.T) {
	if !strings.HasPrefix(slogan, "E") {
		t.Errorf("slogan should start with a capital letter, got %q", slogan)
	}
}

func TestSloganFitsAnEightyColumnTerminal(t *testing.T) {
	if lipgloss.Width(slogan) > 48 {
		t.Errorf("slogan is %d columns wide", lipgloss.Width(slogan))
	}
}

func fakeTools() []*model.Tool {
	return []*model.Tool{
		{Name: "recent", LastUsed: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "never-used"},
		{Name: "seen", LastUsedSeen: true},
	}
}

func TestNeverUsedSortsFirst(t *testing.T) {
	m := &Model{sortField: sortByLastUsed, sortAsc: true}
	rows := m.sorted(fakeTools())

	if rows[0].Name != "never-used" {
		t.Errorf("ascending last-used should put never-used first, got %q", rows[0].Name)
	}
	if rows[len(rows)-1].Name != "recent" {
		t.Errorf("most recently used should be last, got %q", rows[len(rows)-1].Name)
	}
}

func TestStarsFromInstallCounts(t *testing.T) {
	m := &Model{popularity: map[string]int{
		"everyone": 8_000_000,
		"popular":  1_500_000,
		"common":   400_000,
		"niche":    75_000,
		"rare":     12,
	}}
	cases := map[string]int{
		"everyone": 5,
		"popular":  4,
		"common":   3,
		"niche":    2,
		"rare":     1,
		"unknown":  0,
	}
	for name, want := range cases {
		if got := m.stars(&model.Tool{Name: name}); got != want {
			t.Errorf("stars(%s) = %d, want %d", name, got, want)
		}
	}
}

func TestHeaderCursorWalksVisibleColumns(t *testing.T) {
	m := &Model{width: 120, height: 40, headerCursor: sortByName}
	seen := map[sortField]bool{}
	for i := 0; i < len(m.visibleSortFields()); i++ {
		seen[m.headerCursor] = true
		m.moveHeaderCursor(1)
	}
	if len(seen) != len(m.visibleSortFields()) {
		t.Errorf("h visited %d columns, want %d", len(seen), len(m.visibleSortFields()))
	}
	if m.headerCursor != sortByName {
		t.Error("h should wrap back to the first column")
	}
}

func TestShiftHWalksBackwards(t *testing.T) {
	m := &Model{width: 120, height: 40, headerCursor: sortByName}
	m.moveHeaderCursor(-1)

	fields := m.visibleSortFields()
	if want := fields[len(fields)-1]; m.headerCursor != want {
		t.Errorf("H from the first column = %v, want the last (%v)", m.headerCursor, want)
	}

	m.moveHeaderCursor(1)
	if m.headerCursor != sortByName {
		t.Error("h after wrapping backwards should land on the first column again")
	}
}

func TestExportTextHasColumnsAndRows(t *testing.T) {
	m := &Model{filter: filterMine}
	rows := []*model.Tool{
		{Name: "ffmpeg", Version: "8.0.1", SizeBytes: 2048, Source: model.Brew, UserInstalled: true},
	}
	text := m.exportText(rows)

	for _, want := range []string{"PACKAGE", "VERSION", "WHAT IT'S FOR", "SIZE", "POPULAR", "LAST USED", "ffmpeg", "brew", "1 package"} {
		if !strings.Contains(text, want) {
			t.Errorf("export is missing %q", want)
		}
	}
}

func TestUpgradeTargets(t *testing.T) {
	outdated := &model.Tool{ID: "brew:a", Name: "a", Version: "1.0", Latest: "2.0", UserInstalled: true}
	current := &model.Tool{ID: "brew:b", Name: "b", Version: "1.0", UserInstalled: true}
	rows := []*model.Tool{current, outdated}

	m := &Model{}
	if got := m.upgradeTargets(rows); len(got) != 0 {
		t.Error("a highlighted up-to-date package should not be upgraded")
	}

	m.cursor = 1
	if got := m.upgradeTargets(rows); len(got) != 1 || got[0] != outdated {
		t.Error("u should fall back to the highlighted package when nothing is selected")
	}

	m.selected = map[string]bool{"brew:a": true, "brew:b": true}
	if got := m.upgradeTargets(rows); len(got) != 1 || got[0] != outdated {
		t.Errorf("only outdated selections should be upgraded, got %d", len(got))
	}
}

func TestProgressWordingMatchesTheAction(t *testing.T) {
	m := &Model{action: actionUpdate, pending: []*model.Tool{{}, {}, {}}}
	if view := m.workingView(); !strings.Contains(view, "updating 3") {
		t.Errorf("update progress should say updating, got %q", view)
	}

	m.action = actionRemove
	if view := m.workingView(); !strings.Contains(view, "removing 3") {
		t.Errorf("removal progress should say removing, got %q", view)
	}
}
