package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The emblem is braille dot-matrix: Hanuman flying with the Dronagiri mountain
// in his palm. Each braille cell encodes a 2x4 grid of dots, which is how this
// much detail fits in a terminal at all.
const logoWide = `
               ⢀⣤⣶⣶⣶⣆⣀                           ⣰⡆                 ⣰
           ⢀⣴⣿⣷⡟⠉⠁  ⠉⠙⢿⣦         ⣤⡤⣠⣴⡶⣖⣤⡀   ⣀⢙⣯⡶⣖⣨⡿⡢⡀             ⠠⠾⢇⠱⢄⣠
          ⠤⣾⣿⣿⠟⠁       ⠹⣷⡀       ⢫⣾⣿⣽⣿⣿⣾⣟⣂ ⣸⠇⢞⣽⣿⡿⣽⣋⢧⣋⡂          ⢰⣿⣿⣟⣚⣊⢡⣮⣇⣆
                         ⠻⣧       ⣾⢿⣿⣿⣿⣿⣿⣫⠟ ⣿⡆⣘⣛⣻⣼⣷⢋⡵⠶⣦        ⠴⣶⣴⣦⡯⠹⠯⠿⠼⣫⣶⣬⣀⡆
                          ⢻⣷⣄     ⢻⣼⣿⣿⣿⣿⣿⣿⣿ ⠘⣿⡼⡿⣯⣝⢱⣿⣾⣃⣯⣀     ⢠⢾⠿⣋⢫⠽⣴⣾⣿⣿⣆⢺⡟⠿⢿⢿⣶⣄
                  ⢀⣀⣀⣀⣀⣀   ⠙⠻⣿⣤⣤⣀⣀⣀⠻⢿⣿⣭⡟⢿⠿⢿⣷⣄⢬⣭⠄⠚⢿⢸⣿⣿⣿⣿⣿⡇  ⢀⣴⣷⣶⢿⣿⣶⣿⣏⣹⣟⣿⣿⣷⣿⣷⣿⣿⣤⣽⣄⡀
               ⢤⣤⣶⣿⣿⠿⠿⠿⠿⣷⣦⣀   ⠈⠉⠙⠛⠿⣷⣄     ⣀⣙⠿⣷⣌⢀⡀⣳⣄⡙⠛⠿⣾⡭   ⠈⠉⠉⠁⠉⠉⢩⣬⣭⣭⣉⣭⡷⠔⠊⠉⠉ ⠉⠉⠉⠁
                ⠉⠩⣭⣶⣾⡿⠿⠶⢦⣍⡙⠻⠷⠶⠖⠚⣛⡛⠲⢬⣙⢷⡀⣠⡴⣿⣿⣿⣷⡌⣵⣿⣿⣟⣻⣿⡰⣄⢀         ⢼⣿⣯⡍⠉⠉⠁
                  ⠈⠉⠁⠚⠿⣿⣷⣶⣶⣶⣶⡶⠾⠛⠋⠉⠙⠲⢌⠓⣴⡶⠿⠳⠟⢻⢩⢰⣿⣿⣿⣷⣾⣟⡆⣿⠌⣿⣶⣦⣄⡀ ⢀⣴⣾⣷⣭⠉
                        ⠈⠉⠉⠁⢀⣤⣶⣶⡆⣿⣶⢦⡄⢸⣿⣿⣿⣿⣿⣞⣿⡷⣯⢱⣿⣷⣎⠿⣹⡿⢰⣿⣿⣏⣿⣿⣧⣿⣿⣿⣿⠇
      ⣀⣀ ⢀⡀    ⣀⣠⣤⣤⣄⣀  ⣤⣤⡰⣆⣿⣿⣿⣻⣿⣿⣿⢿⣟⢧⠸⠿⣿⡿⠿⢛⣩⡬⣴⠟⣾⣿⣿⣿⣿⡿⠁⠻⢿⡟⢹⣿⣿⣿⣿⡿⠟⠁
   ⢀⣠⣾⣿⣿⢳⣿⣿⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡘⣿⣷⡙⢾⡿⣿⣷⣯⣟⡛⠷⢭⡃⢷⢲⠰⢶⣾⣿⡟⣽⢫⣾⣿⢿⣯⡿⠋
 ⣴⣿⣿⡿⠟⢛⡘⢿⡯⠩⣩⣭⣭⣭⣍⣉⣉⠉⠛⠿⠿⣿⣌⠻⣿⣷⣮⣭⣉⣛⣛⡛⠛⠒⣂⣠⡾⣭⣿⢻⣿⣧⣿⣌⣩⡶⢫⡇
⠺⠿⠋⣡⣴⣾⣿⣿⠧⣆⣇⠿⠿⠿⢿⣿⣿⣿⣿⣿⣶⣆⢰⣶⡆⣦⣭⣭⣭⣭⣤⣴⣶⠾⢟⣽⡌⠳⠿⠞⢒⡛⠛⠛⣉⠔⣴⠋
  ⢰⡿⠻⠋⠁  ⠈⠁      ⠉⠉⠛⠿⠿⠆⢻⣿⣔⠶⠶⠶⣒⣛⣯⣵⠾⠛⠉  ⠐⢿⣦⣍⣛⣩⡴⠞⠁
                         ⠹⢿⣿⣿⡿⠿⠏⠉        ⠈⠉⠉
`

// logoNarrow is the same figure drawn for tighter windows.
const logoNarrow = `
         ⢀⣀⣠⣴⠶⠶⣦⣄⡀         ⢀⣀     ⣀⢀⡀⢶⡀           ⢀⡴⡀
       ⢀⣰⣿⣿⡟    ⠙⢿⣄     ⠰⣿⣾⣿⣽⣷⣦ ⢠⣖⣿⣯⣷⣾⣏⢦        ⢠⣤⡭⣱⠭⣲⢄
       ⠈⠉⠉⠉      ⠈⢿⡀    ⢠⡿⣿⣿⣿⣿⣿⠇⣿⡄⣟⣿⣽⡿⣻⣭⡄     ⢠⣦⣭⣽⢳⣿⣧⢟⣅⣃⡄
                  ⠈⢿⣦⡀  ⠈⢷⣿⣿⣿⢿⣿⣇⠘⢿⠿⢫⣭⣾⣿⣭⣧⡀   ⡰⢟⣫⢩⣱⣾⣿⣷⡸⣝⢟⣿⠶⡄
          ⢀⣀⣤⣶⣶⣶⣶⣦⡀ ⠈⠛⠿⠶⣶⣦⡙⠛⠛⠙⠉⠻⣷⣝⠊⠈⣝⠻⣿⣿⣿⡟  ⠾⠿⠻⠿⢟⣿⣶⠿⣿⣛⡿⠿⠛⠾⠷⠦
           ⠙⠻⣿⣷⡶⢶⣯⣛⠷⢶⣶⠶⢶⡶⢮⣟⢦⣀⣤⣶⣿⣷⣝⣷⣷⣟⣷⣦⣌⡉      ⢰⣾⣿⠛⠛⠉
             ⠈⠁⠚⠿⢷⣷⣶⠶⠶⠛⠋⣉⠓⠮⢃⣶⣟⣱⡙⣭⡸⣿⢿⡷⣿⣇⣿⢹⣷⡦⣄⡀⣠⣶⣷⡍⠁
           ⢀⣀⣀  ⢀⣀⢠⢠⣴⣾⣿⣿⣿⣯⣶⢸⣿⣿⣿⠿⣻⢫⣶⣿⣿⣮⣼⠏⣿⣿⢳⣿⣿⣿⣿⠟⠁
  ⣀⣴⣿⣦⣾⢲⣶⣶⣾⣿⣿⣿⣿⣶⣞⣿⣟⢿⣿⣿⣿⣿⡿⢿⣎⢦⡎⢭⣶⣿⣻⣿⣾⣿⣿⠿⠁  ⠉⠈⠉⠉⠉
⣠⣿⡿⠿⣛⣍⠻⢈⣭⣿⣿⣯⣭⣍⣛⠿⠿⠯⠻⢿⣾⣯⡽⠿⠟⣓⣢⢰⣿⢿⢹⣿⣿⣭⠽⣪⠇
⠉⠁⣴⣿⠿⠛⠙⠼⠛⠛⠛⠛⠻⠿⢿⣿⣞⣿⣯⣒⣒⣒⣲⣾⣟⣭⡾⠃⠙⣭⡙⢬⢥⣒⡽⠁
  ⠉            ⠈ ⠈⠻⣷⣷⡾⠿⠛⠉⠁   ⠈⠙⠛⠋⠁
`

// logoStyle matches the orange of the original artwork (terminal colour 208).
var logoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

// slogan is the one line that explains the whole idea.
const slogan = "Every tool on your Mac, lifted into one place"

// logoLines picks the version of the artwork that fits the window.
func logoLines(width int) []string {
	switch {
	case width >= 86:
		return trimBlank(logoWide)
	case width >= 64:
		return trimBlank(logoNarrow)
	default:
		return nil
	}
}

func trimBlank(art string) []string {
	lines := strings.Split(strings.Trim(art, "\n"), "\n")
	return lines
}

// renderLogo returns the coloured artwork, or an empty string when the window
// is too narrow to do it justice.
func renderLogo(width int) string {
	lines := logoLines(width)
	if len(lines) == 0 {
		return ""
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = logoStyle.Render(line)
	}
	return strings.Join(out, "\n")
}

// Emblem exposes the artwork for the `logo` subcommand.
func Emblem(width int) string {
	return renderLogo(width)
}

// logoWidth is the widest line of the chosen artwork.
func logoWidth(width int) int {
	widest := 0
	for _, line := range logoLines(width) {
		if w := lipgloss.Width(line); w > widest {
			widest = w
		}
	}
	return widest
}
