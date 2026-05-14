package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorBrand      = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#22C55E"}
	colorAccent     = lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#22D3EE"}
	colorAccentSoft = lipgloss.AdaptiveColor{Light: "#67E8F9", Dark: "#0E7490"}
	colorSuccess    = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}
	colorWarn       = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}
	colorError      = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	colorDim        = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#64748B"}
	colorMuted      = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#94A3B8"}
	colorText       = lipgloss.AdaptiveColor{Light: "#0F172A", Dark: "#E2E8F0"}
	colorBorder     = lipgloss.AdaptiveColor{Light: "#CBD5E1", Dark: "#334155"}
	colorBorderHi   = lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#22D3EE"}
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	frameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	frameActiveStyle = frameStyle.
				BorderForeground(colorBorderHi)

	statusConnectedStyle    = lipgloss.NewStyle().Foreground(colorSuccess)
	statusDisconnectedStyle = lipgloss.NewStyle().Foreground(colorDim)
	statusConnectingStyle   = lipgloss.NewStyle().Foreground(colorWarn)
	statusErrorStyle        = lipgloss.NewStyle().Foreground(colorError)

	itemNameStyle      = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	itemDimStyle       = lipgloss.NewStyle().Foreground(colorDim)
	itemDetailStyle    = lipgloss.NewStyle().Foreground(colorMuted)
	itemSelectedBg     = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#1E293B"})
	itemSelectedAccent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	labelStyle = lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
	valueStyle = lipgloss.NewStyle().Foreground(colorText)

	hintStyle = lipgloss.NewStyle().Foreground(colorDim)

	errorStyle = lipgloss.NewStyle().Foreground(colorError).Bold(true)

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(colorWarn).
			Padding(1, 2)

	headerBarStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder).
			Padding(0, 1)

	bannerStyle = lipgloss.NewStyle().
			Foreground(colorBrand).
			Bold(true)

	footerBarStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorBorder).
			Padding(0, 1)
)

const (
	glyphConnected    = "●"
	glyphDisconnected = "○"
	glyphConnecting   = "◐"
	glyphError        = "⚠"
	glyphCursor       = "▸"
	glyphCursorEmpty  = " "
)
