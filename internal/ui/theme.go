package ui

import "github.com/charmbracelet/lipgloss"

// The palette is deliberately quiet: colour marks state, it does not shout.
var (
	cAccent    = lipgloss.Color("#A78BFA")
	cAccentDim = lipgloss.Color("#6D5BD0")
	cText      = lipgloss.Color("#E5E7EB")
	cMuted     = lipgloss.Color("#7A8599")
	cFaint     = lipgloss.Color("#4B5563")
	cGreen     = lipgloss.Color("#4ADE80")
	cRed       = lipgloss.Color("#F87171")
	cYellow    = lipgloss.Color("#FBBF24")
	cCyan      = lipgloss.Color("#67E8F9")
	cOrange    = lipgloss.Color("#FB923C")
	cPink      = lipgloss.Color("#FDA4AF")
	cLine      = lipgloss.Color("#3F3F46")
	cSelected  = lipgloss.Color("#2A2440")
	cGold      = lipgloss.Color("#FBBF24")
	cTab       = lipgloss.Color("#F97316")
	cPanel     = lipgloss.Color("#0B1220")
	cPanelAlt  = lipgloss.Color("#111C33")
	cTabDim    = lipgloss.Color("#C2410C")
)

var (
	styleWordmark = lipgloss.NewStyle().
			Foreground(cTab).
			Bold(true)

	styleTagline = lipgloss.NewStyle().Foreground(cMuted)

	styleNameSub = lipgloss.NewStyle().Foreground(cGold).Bold(true)

	styleMuted = lipgloss.NewStyle().Foreground(cMuted)
	styleFaint = lipgloss.NewStyle().Foreground(cFaint)
	styleText  = lipgloss.NewStyle().Foreground(cText)

	styleTabActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0B0B0F")).
			Background(cTab).
			Bold(true).
			Padding(0, 1)

	styleTabIdle = lipgloss.NewStyle().
			Foreground(cTabDim).
			Background(cPanel).
			Padding(0, 1)

	stylePanel      = lipgloss.NewStyle().Background(cPanel)
	stylePanelMuted = lipgloss.NewStyle().Foreground(cMuted).Background(cPanel)

	styleHeaderCell = lipgloss.NewStyle().
			Foreground(cTab).
			Background(cPanelAlt).
			Bold(true)

	styleWarning = lipgloss.NewStyle().Foreground(cRed).Bold(true)

	styleKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	styleHeaderActive = lipgloss.NewStyle().
				Foreground(cGold).
				Background(lipgloss.Color("#3A2A1A")).
				Bold(true)

	styleRow         = lipgloss.NewStyle().Foreground(cText)
	styleRowSelected = lipgloss.NewStyle().Foreground(cText).Background(cSelected)
	styleCursor      = lipgloss.NewStyle().Foreground(cAccent).Bold(true)

	styleGreen  = lipgloss.NewStyle().Foreground(cGreen)
	styleRed    = lipgloss.NewStyle().Foreground(cRed)
	styleYellow = lipgloss.NewStyle().Foreground(cYellow)
	styleCyan   = lipgloss.NewStyle().Foreground(cCyan)
	styleAccent = lipgloss.NewStyle().Foreground(cAccent)

	styleDetailBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cLine).
			Padding(0, 1)

	styleFooter = lipgloss.NewStyle().Foreground(cMuted)

	styleHelpKey = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	styleHelpBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cAccentDim).
			Padding(1, 2)

	styleDialogue = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cRed).
			Padding(1, 2)

	styleError = lipgloss.NewStyle().Foreground(cRed)
	styleOK    = lipgloss.NewStyle().Foreground(cGreen)
)

// Faint source badges: each manager gets a colour, kept subtle.
func sourceStyle(source string) lipgloss.Style {
	switch source {
	case "brew":
		return lipgloss.NewStyle().Foreground(cYellow)
	case "cask":
		return lipgloss.NewStyle().Foreground(cOrange)
	case "npm":
		return lipgloss.NewStyle().Foreground(cRed)
	case "gem":
		return lipgloss.NewStyle().Foreground(cPink)
	case "cargo":
		return lipgloss.NewStyle().Foreground(cOrange)
	case "pipx":
		return lipgloss.NewStyle().Foreground(cGreen)
	case "go":
		return lipgloss.NewStyle().Foreground(cCyan)
	}
	return styleMuted
}
