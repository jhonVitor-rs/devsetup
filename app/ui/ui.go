package ui

import "github.com/charmbracelet/lipgloss"

// Paleta por papel. Para trocar o visual, mexa só aqui.
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#1F8A5B", Dark: "#25A065"}
	colorOnPrim  = lipgloss.Color("#FFFDF5")
	colorMuted   = lipgloss.AdaptiveColor{Light: "#8A8A8A", Dark: "#6C6C6C"}
	colorSuccess = lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}
	colorWarn    = lipgloss.AdaptiveColor{Light: "#C27C0E", Dark: "#E5A93D"}
	colorError   = lipgloss.AdaptiveColor{Light: "#D12F2F", Dark: "#F25D5D"}
)

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(colorOnPrim).
			Background(colorPrimary).
			Padding(0, 1).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Painéis: a borda muda conforme o foco.
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	PanelFocusedStyle = PanelStyle.
				BorderForeground(colorPrimary)

	PanelTitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Estados de cada linha da lista.
	CursorRowStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	InstalledStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	WarnStyle      = lipgloss.NewStyle().Foreground(colorWarn)
	ErrorStyle     = lipgloss.NewStyle().Foreground(colorError)
)
