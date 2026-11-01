// Package ui renders the interactive terminal application.
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
	"github.com/MasterShivnath007/dronagiri/internal/source"
)

type mode int

const (
	modeLoading mode = iota
	modeList
	modeConfirm
	modeConfirmFinal
	modeWorking
	modeResults
	modeHelp
)

// action is what the program is currently busy doing, so the progress and
// result screens say "updating" or "removing" correctly.
type action int

const (
	actionRemove action = iota
	actionUpdate
)

func (a action) gerund() string {
	if a == actionUpdate {
		return "updating"
	}
	return "removing"
}

func (a action) past() string {
	if a == actionUpdate {
		return "updated"
	}
	return "removed"
}

// sortField is the column the list is ordered by.
type sortField int

const (
	sortByName sortField = iota
	sortBySize
	sortByVersion
	sortByInstalled
	sortByLastUsed
	sortByPopular
	sortBySource
	sortFieldCount
)

func (s sortField) label() string {
	switch s {
	case sortBySize:
		return "size"
	case sortByVersion:
		return "version"
	case sortByInstalled:
		return "install date"
	case sortByLastUsed:
		return "last used"
	case sortByPopular:
		return "popularity"
	case sortBySource:
		return "source"
	default:
		return "name"
	}
}

type filter int

const (
	filterMine filter = iota
	filterDeps
	filterUpdates
	filterBroken
	filterAll
)

func (f filter) label() string {
	switch f {
	case filterMine:
		return "Mine"
	case filterDeps:
		return "Dependencies"
	case filterUpdates:
		return "Updates"
	case filterBroken:
		return "Broken"
	default:
		return "All"
	}
}

func (f filter) match(t *model.Tool) bool {
	switch f {
	case filterMine:
		return t.UserInstalled
	case filterDeps:
		return !t.UserInstalled
	case filterUpdates:
		return t.Outdated()
	case filterBroken:
		return t.Broken
	default:
		return true
	}
}

type result struct {
	tool *model.Tool
	out  string
	err  error
}

type (
	loadedMsg     struct{ inv *model.Inventory }
	refreshedMsg  struct{ inv *model.Inventory }
	resultsMsg    struct{ results []result }
	popularityMsg struct{ counts map[string]int }
)

// Model is the whole application state.
type Model struct {
	runner *runner.Runner
	inv    *model.Inventory

	width, height  int
	cursor, offset int
	filter         filter
	selected       map[string]bool
	query          string
	searching      bool

	sortField    sortField
	sortAsc      bool
	headerCursor sortField
	popularity   map[string]int

	input   textinput.Model
	spinner spinner.Model

	mode    mode
	action  action
	results []result
	pending []*model.Tool
	note    string
}

// New builds the application.
func New(r *runner.Runner) *Model {
	input := textinput.New()
	input.Placeholder = "search packages…"
	input.CharLimit = 64
	input.Prompt = "/ "

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(cTab)

	return &Model{
		runner:       r,
		filter:       filterMine,
		selected:     make(map[string]bool),
		sortField:    sortByName,
		sortAsc:      true,
		headerCursor: sortByName,
		input:        input,
		spinner:      sp,
		mode:         modeLoading,
	}
}

func loadCmd(r *runner.Runner) tea.Cmd {
	return func() tea.Msg { return loadedMsg{inv: source.LoadAll(r, nil)} }
}

// popularityCmd loads install counts from cache when possible, and from
// Homebrew's public analytics in the background when not. It never blocks the
// package list: if it fails, the column simply stays blank.
func popularityCmd() tea.Cmd {
	return func() tea.Msg {
		if counts, fresh := source.CachedPopularity(); fresh && counts != nil {
			return popularityMsg{counts: counts}
		}

		cached, _ := source.CachedPopularity()
		fetched, err := source.FetchPopularity(2 * time.Minute)
		if err != nil {
			return popularityMsg{counts: cached}
		}
		return popularityMsg{counts: fetched}
	}
}

func refreshCmd(r *runner.Runner) tea.Cmd {
	return func() tea.Msg { return refreshedMsg{inv: source.LoadAll(r, nil)} }
}

func uninstallCmd(r *runner.Runner, tools []*model.Tool) tea.Cmd {
	return func() tea.Msg {
		results := make([]result, 0, len(tools))
		for _, tool := range tools {
			out, err := source.Uninstall(r, tool)
			results = append(results, result{tool: tool, out: out, err: err})
		}
		return resultsMsg{results: results}
	}
}

func upgradeCmd(r *runner.Runner, tools []*model.Tool) tea.Cmd {
	return func() tea.Msg {
		results := make([]result, 0, len(tools))
		for _, tool := range tools {
			out, err := source.Upgrade(r, tool)
			results = append(results, result{tool: tool, out: out, err: err})
		}
		return resultsMsg{results: results}
	}
}

// Init starts the first inventory load.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadCmd(m.runner), popularityCmd())
}

// Update handles every message.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case loadedMsg:
		m.inv = msg.inv
		m.mode = modeList
		m.clamp()
		return m, nil

	case refreshedMsg:
		m.inv = msg.inv
		m.clamp()
		return m, nil

	case popularityMsg:
		m.popularity = msg.counts
		return m, nil

	case resultsMsg:
		m.results = msg.results
		m.mode = modeResults
		m.selected = make(map[string]bool)
		return m, refreshCmd(m.runner)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.mode {
	case modeList:
		return m.handleListKey(msg)
	case modeConfirm:
		switch key {
		case "y", "Y", "enter":
			m.mode = modeConfirmFinal
			return m, nil
		case "n", "N", "esc", "q":
			m.mode = modeList
			m.pending = nil
			return m, nil
		}
		return m, nil
	case modeConfirmFinal:
		switch key {
		case "y", "Y":
			m.mode = modeWorking
			return m, tea.Batch(m.spinner.Tick, uninstallCmd(m.runner, m.pending))
		case "n", "N", "esc", "enter", "q":
			m.mode = modeConfirm
			return m, nil
		}
		return m, nil
	case modeResults:
		switch key {
		case "q":
			return m, tea.Quit
		case "esc", "enter", " ":
			m.mode = modeList
			m.results = nil
			return m, nil
		}
		return m, nil
	case modeHelp:
		m.mode = modeList
		return m, nil
	case modeWorking:
		return m, nil
	}
	return m, nil
}

func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.searching {
		switch key {
		case "enter":
			m.searching = false
			m.input.Blur()
			return m, nil
		case "esc":
			m.searching = false
			m.input.Blur()
			m.input.SetValue("")
			m.query = ""
			m.cursor, m.offset = 0, 0
			return m, nil
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.query = m.input.Value()
			m.cursor, m.offset = 0, 0
			return m, cmd
		}
	}

	rows := m.rows()
	page := m.visibleRows()

	if key != "s" && key != "S" && key != "h" && key != "H" {
		m.note = ""
	}

	switch key {
	case "q":
		return m, tea.Quit
	case "up", "k":
		m.cursor--
	case "down", "j":
		m.cursor++
	case "pgup":
		m.cursor -= page
	case "pgdown":
		m.cursor += page
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = len(rows) - 1
	case " ":
		if len(rows) > 0 && m.cursor < len(rows) {
			tool := rows[m.cursor]
			m.selected[tool.ID] = !m.selected[tool.ID]
			m.cursor++
		}
	case "a":
		m.toggleAll(rows)
	case "tab":
		m.filter = (m.filter + 1) % 5
		m.cursor, m.offset = 0, 0
	case "shift+tab":
		m.filter = (m.filter + 4) % 5
		m.cursor, m.offset = 0, 0
	case "/":
		m.searching = true
		m.input.Focus()
		return m, textinput.Blink
	case "esc":
		m.query = ""
		m.input.SetValue("")
	case "r":
		m.mode = modeLoading
		return m, loadCmd(m.runner)
	case "?":
		m.mode = modeHelp
	case "d":
		targets := m.targets(rows)
		if len(targets) > 0 {
			m.pending = targets
			m.action = actionRemove
			m.mode = modeConfirm
		}
	case "e":
		m.exportCurrentView()
	case "h":
		m.moveHeaderCursor(1)
	case "H":
		m.moveHeaderCursor(-1)
	case "s", "S":
		if m.sortField == m.headerCursor {
			m.sortAsc = !m.sortAsc
		} else {
			m.sortField = m.headerCursor
			m.sortAsc = true
		}
		m.rewind()
		m.note = "sorted by " + m.sortField.label() + " " + m.sortArrow()
	case "u":
		targets := m.upgradeTargets(rows)
		if len(targets) == 0 {
			if len(rows) > 0 && m.cursor >= 0 && m.cursor < len(rows) {
				if !rows[m.cursor].Outdated() {
					m.note = rows[m.cursor].Name + " is already up to date"
				} else {
					m.note = "nothing here has an update available"
				}
			}
			return m, nil
		}
		m.pending = targets
		m.action = actionUpdate
		m.mode = modeWorking
		return m, tea.Batch(m.spinner.Tick, upgradeCmd(m.runner, targets))
	}

	m.clamp()
	return m, nil
}

// targets returns the selected tools, or the hovered tool when nothing is
// selected.
func (m *Model) targets(rows []*model.Tool) []*model.Tool {
	var targets []*model.Tool
	for _, tool := range rows {
		if m.selected[tool.ID] {
			targets = append(targets, tool)
		}
	}
	if len(targets) == 0 && m.cursor < len(rows) {
		targets = append(targets, rows[m.cursor])
	}
	return targets
}

func (m *Model) toggleAll(rows []*model.Tool) {
	allSelected := len(rows) > 0
	for _, tool := range rows {
		if !m.selected[tool.ID] {
			allSelected = false
			break
		}
	}
	for _, tool := range rows {
		m.selected[tool.ID] = !allSelected
	}
}

func (m *Model) rows() []*model.Tool {
	if m.inv == nil {
		return nil
	}
	query := strings.ToLower(strings.TrimSpace(m.query))
	var out []*model.Tool
	for _, tool := range m.inv.Tools {
		if !m.filter.match(tool) {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(tool.Name + " " + tool.Summary)
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		out = append(out, tool)
	}
	return m.sorted(out)
}

// sorted orders the visible rows by the column the user picked. Reversing is
// done by swapping the operands, so ties stay grouped.
func (m *Model) sorted(tools []*model.Tool) []*model.Tool {
	sort.SliceStable(tools, func(i, j int) bool {
		a, b := tools[i], tools[j]
		if !m.sortAsc {
			a, b = b, a
		}

		switch m.sortField {
		case sortBySize:
			return a.SizeBytes < b.SizeBytes
		case sortByVersion:
			return model.CompareVersions(a.Version, b.Version) < 0
		case sortByInstalled:
			return a.InstalledAt.Before(b.InstalledAt)
		case sortByLastUsed:
			return lastUsedKey(a).Before(lastUsedKey(b))
		case sortByPopular:
			return m.stars(a) < m.stars(b)
		case sortBySource:
			return a.Source.Label() < b.Source.Label()
		default:
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	})
	return tools
}

// lastUsedKey puts "never used" first when sorting ascending, which is what
// someone hunting for dead weight actually wants to see.
func lastUsedKey(t *model.Tool) time.Time {
	switch {
	case !t.LastUsed.IsZero():
		return t.LastUsed
	case t.LastUsedSeen:
		return time.Unix(0, 1)
	default:
		return time.Time{}
	}
}

// visibleSortFields is the order the H key walks: only columns that are
// actually on screen.
func (m *Model) visibleSortFields() []sortField {
	w := m.columns()
	fields := []sortField{sortByName, sortByVersion, sortBySize}
	if w.showPopular {
		fields = append(fields, sortByPopular)
	}
	if w.showInstalled {
		fields = append(fields, sortByInstalled)
	}
	if w.showLastUsed {
		fields = append(fields, sortByLastUsed)
	}
	return append(fields, sortBySource)
}

func (m *Model) moveHeaderCursor(delta int) {
	fields := m.visibleSortFields()
	if len(fields) == 0 {
		return
	}

	index := 0
	for i, field := range fields {
		if field == m.headerCursor {
			index = i
			break
		}
	}

	next := ((index+delta)%len(fields) + len(fields)) % len(fields)
	m.headerCursor = fields[next]
	m.note = "column: " + m.headerCursor.label() + "  —  press s to sort"
}

// stars turns a yearly install count into a five-point popularity score.
func (m *Model) stars(tool *model.Tool) int {
	switch count := m.popularity[tool.Name]; {
	case count >= 2_000_000:
		return 5
	case count >= 500_000:
		return 4
	case count >= 100_000:
		return 3
	case count >= 10_000:
		return 2
	case count > 0:
		return 1
	default:
		return 0
	}
}

// upgradeTargets picks what `u` acts on: selected packages that actually have
// an update, or the hovered package when nothing is selected.
func (m *Model) upgradeTargets(rows []*model.Tool) []*model.Tool {
	var targets []*model.Tool
	for _, tool := range rows {
		if m.selected[tool.ID] && tool.Outdated() {
			targets = append(targets, tool)
		}
	}
	if len(targets) == 0 && m.cursor >= 0 && m.cursor < len(rows) && rows[m.cursor].Outdated() {
		targets = append(targets, rows[m.cursor])
	}
	return targets
}

// exportCurrentView writes exactly what is on screen to ~/Downloads.
func (m *Model) exportCurrentView() {
	rows := m.rows()

	home, err := os.UserHomeDir()
	if err != nil {
		m.note = "could not find your home folder"
		return
	}
	dir := filepath.Join(home, "Downloads")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		m.note = "could not create ~/Downloads"
		return
	}

	name := "dronagiri-" + time.Now().Format("2006-01-02-1504") + ".txt"
	path := filepath.Join(dir, name)

	body := m.exportText(rows)

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		m.note = "could not write " + name
		return
	}
	m.note = "exported " + name + " to Downloads"
}

// exportText renders the current view as a plain, aligned text table.
func (m *Model) exportText(rows []*model.Tool) string {
	var total int64
	for _, tool := range rows {
		total += tool.SizeBytes
	}

	var body strings.Builder
	body.WriteString("Dronagiri — installed packages\n")
	body.WriteString("Generated " + time.Now().Format("2 Jan 2006 15:04") + "\n")
	body.WriteString(fmt.Sprintf("View: %s · %d packages · %s\n\n",
		m.filter.label(), len(rows), model.HumanBytes(total)))

	table := tabwriter.NewWriter(&body, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "PACKAGE\tVERSION\tWHAT IT'S FOR\tSIZE\tPOPULAR\tINSTALLED\tLAST USED\tSOURCE")
	for _, tool := range rows {
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			tool.Name,
			tool.Version,
			tool.Summary,
			model.HumanBytes(tool.SizeBytes),
			m.plainStars(tool),
			model.HumanDate(tool.InstalledAt),
			m.plainLastUsed(tool),
			tool.Source.Label(),
		)
	}
	_ = table.Flush()
	return body.String()
}

func (m *Model) plainStars(tool *model.Tool) string {
	stars := m.stars(tool)
	if stars == 0 {
		return "—"
	}
	return strings.Repeat("★", stars) + strings.Repeat("☆", 5-stars)
}

func (m *Model) plainLastUsed(tool *model.Tool) string {
	switch {
	case !tool.LastUsed.IsZero():
		return model.HumanDate(tool.LastUsed)
	case tool.LastUsedSeen:
		return "used"
	default:
		return "never"
	}
}

func (m *Model) sortArrow() string {
	if m.sortAsc {
		return "▲"
	}
	return "▼"
}

func (m *Model) rewind() {
	m.cursor, m.offset = 0, 0
}

func (m *Model) clamp() {
	rows := m.rows()
	if m.cursor >= len(rows) {
		m.cursor = len(rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	page := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+page {
		m.offset = m.cursor - page + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) visibleRows() int {
	if m.height <= 0 {
		return 10
	}
	if m.showDetail() {
		return max(3, m.height-15)
	}
	return max(3, m.height-7)
}

func (m *Model) showDetail() bool {
	return m.height >= 26
}

// View renders the screen.
func (m *Model) View() string {
	if m.width <= 0 {
		return "starting…"
	}
	if m.mode == modeLoading {
		return m.splashView()
	}

	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(m.tabs())
	b.WriteString("\n")

	switch m.mode {
	case modeList:
		b.WriteString(m.table())
		if m.showDetail() {
			b.WriteString("\n")
			b.WriteString(m.detail())
		}
		b.WriteString("\n")
		b.WriteString(m.footer())
	case modeConfirm:
		b.WriteString(m.confirmView())
	case modeConfirmFinal:
		b.WriteString(m.confirmFinalView())
	case modeWorking:
		b.WriteString(m.workingView())
	case modeResults:
		b.WriteString(m.resultsView())
	case modeHelp:
		b.WriteString(m.helpView())
	}
	return b.String()
}

// header is a single line while browsing, so the package list keeps its room.
// The full emblem appears on the splash screen instead.
func (m *Model) header() string {
	left := styleWordmark.Render("Dronagiri") +
		styleNameSub.Render(" package manager") +
		styleFaint.Render("  ·  ") +
		styleTagline.Render(m.tagline())
	stats := m.stats()
	if stats == "" {
		return left
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(stats) - 1
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + stats
}

// splashView is the full-screen welcome: the Hanuman emblem, the name, the
// slogan, and a spinner while every package manager is read.
func (m *Model) splashView() string {
	var b strings.Builder

	if art := renderLogo(m.width); art != "" {
		lines := strings.Split(art, "\n")
		left := (m.width - logoWidth(m.width)) / 2
		if left < 0 {
			left = 0
		}
		top := (m.height - len(lines) - 7) / 2
		for i := 0; i < top; i++ {
			b.WriteString("\n")
		}
		for _, line := range lines {
			b.WriteString(strings.Repeat(" ", left) + line + "\n")
		}
		b.WriteString("\n")
	}

	centre := func(line string) string {
		pad := (m.width - lipgloss.Width(line)) / 2
		if pad < 0 {
			pad = 0
		}
		return strings.Repeat(" ", pad) + line
	}

	b.WriteString(centre(styleWordmark.Render("Dronagiri")+styleNameSub.Render(" package manager")) + "\n")
	b.WriteString(centre(styleTagline.Render(slogan)) + "\n")
	b.WriteString("\n")
	b.WriteString(centre(m.spinner.View() + styleMuted.Render("  reading your package managers…")))
	return b.String()
}

// tagline shortens on narrow terminals so the banner never wraps.
func (m *Model) tagline() string {
	if m.width > 0 && m.width < 110 {
		return "every tool, one mountain"
	}
	return slogan
}

func (m *Model) stats() string {
	if m.inv == nil {
		return ""
	}

	outdated := 0
	for _, tool := range m.inv.Tools {
		if tool.Outdated() {
			outdated++
		}
	}

	extras := []string{
		fmt.Sprintf("%d packages", len(m.inv.Tools)),
		model.HumanBytes(m.inv.TotalSize),
	}
	if outdated > 0 {
		extras = append(extras, styleYellow.Render(fmt.Sprintf("%d outdated", outdated)))
	}
	if len(m.inv.Errors) > 0 {
		extras = append(extras, styleRed.Render(fmt.Sprintf("%d managers failed", len(m.inv.Errors))))
	}
	return styleMuted.Render(strings.Join(extras, " · "))
}

func (m *Model) tabs() string {
	counts := map[filter]int{}
	if m.inv != nil {
		for _, tool := range m.inv.Tools {
			for _, f := range []filter{filterMine, filterDeps, filterUpdates, filterBroken, filterAll} {
				if f.match(tool) {
					counts[f]++
				}
			}
		}
	}

	var parts []string
	for _, f := range []filter{filterMine, filterDeps, filterUpdates, filterBroken, filterAll} {
		label := fmt.Sprintf("%s %d", f.label(), counts[f])
		if m.filter == f {
			parts = append(parts, styleTabActive.Render(label))
		} else {
			parts = append(parts, styleTabIdle.Render(label))
		}
	}

	line := stylePanel.Render(" ") + strings.Join(parts, stylePanel.Render(" "))
	if m.width >= 100 {
		line += stylePanel.Render("   ") + stylePanelMuted.Render("press tab to switch")
	}
	return m.panelLine(line)
}

// panelLine pads a line so its background reaches the right edge, which is
// what makes the tab bar and the table read as one attached surface.
func (m *Model) panelLine(line string) string {
	width := m.width - 1
	if gap := width - lipgloss.Width(line); gap > 0 {
		line += stylePanel.Render(strings.Repeat(" ", gap))
	}
	return line
}

type columnWidths struct {
	name, version, size, what, popular, installed, lastUsed, source int
	showWhat, showPopular, showInstalled, showLastUsed              bool
}

func (m *Model) columns() columnWidths {
	w := columnWidths{version: 14, size: 9, popular: 10, installed: 11, lastUsed: 11, source: 8}
	// Narrow terminals drop columns from the middle rather than letting the
	// last one fall off the edge.
	w.showWhat = m.width >= 104
	w.showPopular = m.width >= 92
	w.showInstalled = m.width >= 84
	w.showLastUsed = m.width >= 100

	columns := 4 // name, version, size, source
	if w.showWhat {
		columns++
	}
	if w.showPopular {
		columns++
	}
	if w.showInstalled {
		columns++
	}
	if w.showLastUsed {
		columns++
	}

	// cursor + checkbox + one separator per column + fixed widths
	fixed := 2 + 4 + columns + w.version + w.size + w.source
	if w.showPopular {
		fixed += w.popular
	}
	if w.showInstalled {
		fixed += w.installed
	}
	if w.showLastUsed {
		fixed += w.lastUsed
	}

	remaining := m.width - fixed - 1
	if remaining < 18 {
		remaining = 18
	}

	if w.showWhat {
		w.name = remaining * 50 / 100
		if w.name < 16 {
			w.name = 16
		}
		w.what = remaining - w.name
		if w.what < 16 {
			w.what = 16
			w.name = max(16, remaining-w.what)
		}
	} else {
		w.name = remaining
		if w.name > 40 {
			w.name = 40
		}
	}
	return w
}

func (m *Model) tableHeader(w columnWidths) string {
	cell := func(text string, width int, right bool, field sortField) string {
		if m.sortField == field {
			switch {
			case lipgloss.Width(text)+2 <= width:
				text += " " + m.sortArrow()
			case lipgloss.Width(text)+1 <= width:
				text += m.sortArrow()
			}
		}
		style := styleHeaderCell
		if m.headerCursor == field {
			style = styleHeaderActive
		}
		if right {
			return style.Render(padLeft(text, width))
		}
		return style.Render(pad(text, width))
	}

	plain := func(text string, width int) string {
		return styleHeaderCell.Render(pad(text, width))
	}

	parts := []string{
		"  " + strings.Repeat(" ", 3),
		cell("PACKAGE", w.name, false, sortByName),
		cell("VERSION", w.version, false, sortByVersion),
	}
	if w.showWhat {
		parts = append(parts, plain("WHAT IT'S FOR", w.what))
	}
	parts = append(parts, cell("SIZE", w.size, true, sortBySize))
	if w.showPopular {
		parts = append(parts, cell("POPULAR", w.popular, false, sortByPopular))
	}
	if w.showInstalled {
		parts = append(parts, cell("INSTALLED", w.installed, false, sortByInstalled))
	}
	if w.showLastUsed {
		parts = append(parts, cell("LAST USED", w.lastUsed, false, sortByLastUsed))
	}
	parts = append(parts, cell("SOURCE", w.source, false, sortBySource))

	sep := lipgloss.NewStyle().Background(cPanelAlt).Render(" ")
	return m.panelLineAlt(strings.Join(parts, sep))
}

// panelLineAlt pads a header line with the alternate panel shade.
func (m *Model) panelLineAlt(line string) string {
	width := m.width - 1
	if gap := width - lipgloss.Width(line); gap > 0 {
		line += lipgloss.NewStyle().Background(cPanelAlt).Render(strings.Repeat(" ", gap))
	}
	return line
}

func (m *Model) table() string {
	w := m.columns()
	var b strings.Builder
	b.WriteString(m.tableHeader(w))

	rows := m.rows()
	page := m.visibleRows()
	end := m.offset + page
	if end > len(rows) {
		end = len(rows)
	}

	for i := m.offset; i < end; i++ {
		b.WriteString("\n")
		b.WriteString(m.row(rows[i], i == m.cursor, w))
	}
	for i := end - m.offset; i < page; i++ {
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Model) row(tool *model.Tool, hovered bool, w columnWidths) string {
	// The hovered row is painted as one solid block: every segment carries the
	// same background so there are no gaps between columns.
	base := cPanel
	if hovered {
		base = cSelected
	}
	paint := func(style lipgloss.Style) lipgloss.Style {
		return style.Background(base)
	}
	sep := paint(lipgloss.NewStyle()).Render(" ")

	cursor := paint(styleRow).Render("  ")
	if hovered {
		cursor = paint(styleCursor).Render("▸ ")
	}

	check := paint(styleFaint).Render("[ ]")
	if m.selected[tool.ID] {
		check = paint(styleGreen).Render("[x]")
	}

	name := pad(truncate(tool.Name, w.name), w.name)
	switch {
	case hovered:
		name = paint(styleCursor).Render(name)
	case m.selected[tool.ID]:
		name = paint(styleText.Bold(true)).Render(name)
	default:
		name = paint(styleRow).Render(name)
	}

	version := truncate(tool.Version, w.version)
	if tool.Outdated() {
		version = truncate(tool.Version+" → "+tool.Latest, w.version)
		version = paint(styleCyan).Render(pad(version, w.version))
	} else {
		version = paint(styleMuted).Render(pad(version, w.version))
	}

	size := paint(styleMuted).Render(padLeft(model.HumanBytes(tool.SizeBytes), w.size))

	line := cursor + check + sep + name + sep + version
	if w.showWhat {
		line += sep + m.noteText(tool, w.what, paint)
	}
	line += sep + size
	if w.showPopular {
		line += sep + m.popularityCell(tool, w.popular, paint)
	}
	if w.showInstalled {
		line += sep + paint(styleMuted).Render(pad(model.HumanDateShort(tool.InstalledAt), w.installed))
	}
	if w.showLastUsed {
		line += sep + m.lastUsedCell(tool, w.lastUsed, paint)
	}
	line += sep + paint(sourceStyle(string(tool.Source))).Render(pad(tool.Source.Label(), w.source))

	// Fill to the right edge so every row shares one continuous surface.
	if gap := m.width - 1 - lipgloss.Width(line); gap > 0 {
		line += paint(lipgloss.NewStyle()).Render(strings.Repeat(" ", gap))
	}
	return line
}

// popularityCell shows how widely a tool is installed across Homebrew users,
// or a dash when no data is available (other package managers, offline runs).
func (m *Model) popularityCell(tool *model.Tool, width int, paint func(lipgloss.Style) lipgloss.Style) string {
	stars := m.stars(tool)
	if stars == 0 {
		return paint(styleFaint).Render(pad("—", width))
	}

	text := strings.Repeat("★", stars) + strings.Repeat("☆", 5-stars)
	style := styleMuted
	switch {
	case stars >= 4:
		style = styleGreen
	case stars == 3:
		style = styleYellow
	}
	return paint(style).Render(pad(text, width))
}

// lastUsedCell reports the strongest evidence we have, and never dresses up a
// guess as a fact: a real date only when Spotlight or a timestamped history
// gave us one, "used" when the command appears in history without dates, and
// "never" when we found no trace at all.
func (m *Model) lastUsedCell(tool *model.Tool, width int, paint func(lipgloss.Style) lipgloss.Style) string {
	switch {
	case !tool.LastUsed.IsZero():
		return paint(styleMuted).Render(pad(model.HumanDateShort(tool.LastUsed), width))
	case tool.LastUsedSeen:
		return paint(styleMuted).Render(pad("used", width))
	default:
		return paint(styleYellow).Render(pad("never", width))
	}
}

// noteText is the one column that answers "why is this on my Mac?": what the
// tool does, what needs it, or what is wrong with it.
func (m *Model) noteText(tool *model.Tool, width int, paint func(lipgloss.Style) lipgloss.Style) string {
	text := tool.Summary
	style := styleMuted

	switch {
	case tool.Broken:
		text, style = "files missing — leftover entry", styleRed
	case !tool.UserInstalled:
		if names := m.inv.RequiredBy[tool.ID]; len(names) > 0 {
			text, style = "needed by "+strings.Join(names, ", "), styleFaint
		} else {
			text, style = "unused dependency — safe to remove", styleYellow
		}
	}

	if strings.TrimSpace(text) == "" {
		text = "—"
	}
	return paint(style).Render(pad(truncate(text, width), width))
}

func (m *Model) detail() string {
	rows := m.rows()
	if len(rows) == 0 || m.cursor >= len(rows) || m.cursor < 0 {
		return ""
	}
	tool := rows[m.cursor]
	width := m.width - 6
	if width < 20 {
		width = 20
	}

	title := styleAccent.Bold(true).Render(tool.Name) + "  " +
		styleMuted.Render(tool.Version) + "  " +
		sourceStyle(string(tool.Source)).Render(tool.Source.Label())

	var lines []string
	lines = append(lines, title)

	summary := tool.Summary
	if summary == "" {
		summary = "No description available."
	}
	lines = append(lines, styleText.Render(truncate(summary, width)))

	meta := []string{model.HumanBytes(tool.SizeBytes), "installed " + model.HumanDate(tool.InstalledAt)}
	if tool.Outdated() {
		meta = append(meta, styleCyan.Render("update available: "+tool.Latest))
	}
	lines = append(lines, styleMuted.Render(strings.Join(meta, "  ·  ")))
	lines = append(lines, lastUsedDetail(tool, width))

	if tool.Path != "" {
		lines = append(lines, styleFaint.Render(truncate(tool.Path, width)))
	}

	switch {
	case tool.Broken:
		lines = append(lines, styleRed.Render("✗ files missing · ")+styleMuted.Render("the manager still lists it, but nothing is on disk"))
	case tool.Protected:
		lines = append(lines, styleRed.Render("protected · ")+styleMuted.Render(tool.ProtectReason))
	case tool.IsDependency():
		names := m.inv.RequiredBy[tool.ID]
		if len(names) == 0 {
			lines = append(lines, styleFaint.Render("dependency · nothing currently requires it"))
		} else {
			lines = append(lines, styleMuted.Render("dependency of: ")+styleGreen.Render(truncate(strings.Join(names, ", "), width-15)))
		}
	}

	return styleDetailBox.Width(width).Render(strings.Join(lines, "\n"))
}

// lastUsedDetail explains where the last-used evidence came from, because the
// difference between "never" and "we could not check" matters.
func lastUsedDetail(tool *model.Tool, width int) string {
	switch {
	case !tool.LastUsed.IsZero():
		return styleMuted.Render(truncate(
			"last used "+model.HumanDate(tool.LastUsed)+" · from "+tool.LastUsedSource, width))
	case tool.LastUsedSeen:
		return styleMuted.Render(truncate(
			"last used · seen in shell history, no date recorded", width))
	default:
		return styleYellow.Render(truncate(
			"last used · no trace in your shell history (scripts and editors are not tracked)", width))
	}
}

func (m *Model) footer() string {
	var selected int
	var size int64
	for _, tool := range m.invTools() {
		if m.selected[tool.ID] {
			selected++
			size += tool.SizeBytes
		}
	}

	hints := []struct{ key, label string }{
		{"↑↓", "move"},
		{"space", "select"},
		{"/", "search"},
		{"tab", "views"},
		{"h", "column"},
		{"s", "sort"},
		{"u", "upgrade"},
		{"?", "help"},
		{"e", "export"},
		{"q", "quit"},
	}
	var hintLine strings.Builder
	hintLine.WriteString(" ")
	for i, hint := range hints {
		if i > 0 {
			hintLine.WriteString(styleFooter.Render(" · "))
		}
		hintLine.WriteString(styleKey.Render(hint.key))
		hintLine.WriteString(styleFooter.Render(" " + hint.label))
	}
	left := hintLine.String()
	if selected > 0 {
		badge := styleGreen.Render(fmt.Sprintf(" %d selected · %s ", selected, model.HumanBytes(size)))
		left = badge + "  " + styleWarning.Render("d to remove") + styleFooter.Render(" · space to clear")
	}
	if m.searching {
		return " " + m.input.View()
	}
	if m.note != "" {
		return " " + styleYellow.Render(m.note)
	}
	return left
}

func (m *Model) invTools() []*model.Tool {
	if m.inv == nil {
		return nil
	}
	return m.inv.Tools
}

// dialogue draws a centred, bordered panel — removal is destructive, so it
// gets a proper window rather than another line of text.
func (m *Model) dialogue(title string, body []string) string {
	width := m.width - 14
	if width > 74 {
		width = 74
	}
	if width < 30 {
		width = 30
	}

	box := styleDialogue.Width(width).Render(title + "\n\n" + strings.Join(body, "\n"))

	prefix := strings.Repeat(" ", max(0, (m.width-lipgloss.Width(box))/2))
	lines := strings.Split(box, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func (m *Model) confirmView() string {
	var size int64
	var risky, protected []string

	body := make([]string, 0, len(m.pending)+6)
	for _, tool := range m.pending {
		size += tool.SizeBytes
		if tool.Protected {
			protected = append(protected, tool.Name)
		} else if tool.IsDependency() {
			risky = append(risky, tool.Name)
		}
		body = append(body, styleText.Render(pad(truncate(tool.Name, 34), 34))+
			styleMuted.Render(padLeft(model.HumanBytes(tool.SizeBytes), 10)))
	}

	body = append(body, "")
	body = append(body, styleMuted.Render("This frees "+model.HumanBytes(size)+". It cannot be undone."))

	if len(risky) > 0 {
		body = append(body, "")
		body = append(body, styleYellow.Render("! Dependencies of other tools: "+truncate(strings.Join(risky, ", "), 50)))
	}
	if len(protected) > 0 {
		body = append(body, "")
		body = append(body, styleRed.Render("✗ Protected, will be skipped: "+truncate(strings.Join(protected, ", "), 50)))
	}

	body = append(body, "")
	body = append(body, styleFaint.Render("step 1 of 2")+"   "+
		styleGreen.Render("y")+styleMuted.Render(" continue")+"    "+
		styleRed.Render("n")+styleMuted.Render(" cancel"))

	return m.dialogue(styleWarning.Render(fmt.Sprintf("Remove %d package(s)?", len(m.pending))), body)
}

func (m *Model) confirmFinalView() string {
	var size int64
	body := make([]string, 0, len(m.pending)+6)

	for _, tool := range m.pending {
		size += tool.SizeBytes
		body = append(body, styleRed.Render("✗ ")+styleText.Render(tool.Name))
	}

	body = append(body, "")
	body = append(body, styleText.Render(fmt.Sprintf("You are about to uninstall %d package(s) and free %s.",
		len(m.pending), model.HumanBytes(size))))
	body = append(body, "")
	body = append(body, styleRed.Render("This cannot be undone."))
	body = append(body, "")
	body = append(body, styleFaint.Render("step 2 of 2")+"   "+
		styleGreen.Render("y")+styleMuted.Render(" uninstall now")+"    "+
		styleRed.Render("n")+styleMuted.Render(" go back"))

	return m.dialogue(styleWarning.Render("⚠   Last chance"), body)
}

func (m *Model) workingView() string {
	return "  " + m.spinner.View() + styleMuted.Render(
		fmt.Sprintf("  %s %d package(s) — this can take a while…", m.action.gerund(), len(m.pending)))
}

func (m *Model) resultsView() string {
	var b strings.Builder
	var ok, failed int
	for _, res := range m.results {
		if res.err != nil {
			failed++
		} else {
			ok++
		}
	}

	head := styleGreen.Bold(true).Render(fmt.Sprintf("  ✓ %d %s", ok, m.action.past()))
	if failed > 0 {
		head += styleRed.Bold(true).Render(fmt.Sprintf("   ✗ %d failed", failed))
	}
	b.WriteString(head + "\n\n")

	for _, res := range m.results {
		if res.err != nil {
			b.WriteString("  " + styleRed.Render("✗ ") + styleText.Render(res.tool.Name) + "\n")
			b.WriteString("      " + styleMuted.Render(truncate(firstLine(res.err.Error()), m.width-8)) + "\n")
			continue
		}
		b.WriteString("  " + styleGreen.Render("✓ ") + styleText.Render(res.tool.Name) + "\n")
	}

	b.WriteString("\n" + styleMuted.Render("  press enter to continue"))
	return b.String()
}

func (m *Model) helpView() string {
	lines := [][2]string{
		{"↑ ↓  j k", "move between packages"},
		{"space", "select or deselect a package"},
		{"a", "select everything in the current view"},
		{"d", "uninstall the selected packages (asks twice)"},
		{"u", "update the selected (or highlighted) package"},
		{"/", "search by name or description"},
		{"tab", "switch view: Mine, Dependencies, Updates, Broken, All"},
		{"e", "export this view to ~/Downloads"},
		{"h", "move the header highlight forward"},
		{"H", "move the header highlight backward"},
		{"s", "sort by the highlighted column (again to reverse)"},
		{"r", "reload from the package managers"},
		{"e", "export"},
		{"q", "quit"},
	}

	var b strings.Builder
	b.WriteString(styleAccent.Bold(true).Render("Dronagiri package manager") +
		styleMuted.Render("  ·  keybindings") + "\n\n")
	for _, line := range lines {
		b.WriteString("  " + styleHelpKey.Render(pad(line[0], 12)) + styleText.Render(line[1]) + "\n")
	}
	b.WriteString("\n" + styleMuted.Render("  press any key to close"))
	return styleHelpBox.Render(b.String())
}

// MARK: - small helpers

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func pad(s string, width int) string {
	length := lipgloss.Width(s)
	if length == width {
		return s
	}
	if length > width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-length)
}

func padLeft(s string, width int) string {
	length := lipgloss.Width(s)
	if length == width {
		return s
	}
	if length > width {
		return truncate(s, width)
	}
	return strings.Repeat(" ", width-length) + s
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
