package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Dark Mode Color Palette with Pastel Pink
const (
	// Primary pink tones
	PastelPink     = "#FFB3D1"
	SoftPink       = "#FFCCE0"
	DeepPink       = "#FF99CC"
	BlushPink      = "#FFE6F2"
	DarkBackground = "#1E1E1E"
	OffWhite       = "#F2F2F2"

	// Complementary colors
	MintGreen      = "#B3FFD1"
	LavenderPurple = "#D1B3FF"
	SoftCream      = "#FFF9FC"
	MidGray        = "#AAAAAA"
	CharcoalGray   = "#4A4A4A"
	SoftGray       = "#8A8A8A"
)

// Style definitions using lipgloss
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CharcoalGray)).
			Background(lipgloss.Color(SoftCream))

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DarkBackground)).
			Width(40).
			Align(lipgloss.Left).
			Bold(true).
			Padding(1, 2)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DarkBackground)).
			PaddingLeft(1).
			PaddingRight(1).
			MarginBottom(0).
			Width(30)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(DarkBackground)).
				Background(lipgloss.Color(BlushPink)).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1).
				MarginBottom(0).
				Width(30)

	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PastelPink)).
			Bold(true)

	CountBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DarkBackground)).
			Background(lipgloss.Color(MintGreen)).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(MintGreen))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DarkBackground)).
			Background(lipgloss.Color(MintGreen)).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(MintGreen))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF6B9D")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B9D"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MidGray)).
			Italic(true).
			MarginTop(1)

	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PastelPink))

	FileItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DarkBackground)).
			PaddingLeft(2).
			PaddingRight(2)

	SelectedFileItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(DarkBackground)).
				Background(lipgloss.Color(MintGreen)).
				Bold(true).
				PaddingLeft(2).
				PaddingRight(2).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#A1E9C5"))

	BackOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MidGray)).
			Italic(true).
			PaddingLeft(2).
			PaddingRight(2)

	SelectedBackOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(DarkBackground)).
				Background(lipgloss.Color(MidGray)).
				Bold(true).
				Italic(true).
				PaddingLeft(2).
				PaddingRight(2).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#666666"))
)

func RenderMenuTitle(text string) string {
	return TitleStyle.Render(text)
}

func RenderAppHeader() string {
	return TitleStyle.Render("FitTkr CLI")
}

func RenderCountBadge(count int) string {
	if count < 0 {
		// Reserve space using "00" width for alignment
		width := lipgloss.Width(CountBadgeStyle.Render("00"))
		return lipgloss.NewStyle().Width(width).Render(" ")
	}
	return CountBadgeStyle.Render(fmt.Sprintf("%d", count))
}

func RenderMenuItem(text string, isSelected bool, count int) string {
	countBadge := RenderCountBadge(count)
	var styledText string

	if isSelected {
		cursor := CursorStyle.Render("❯ ")
		styledText = SelectedMenuItemStyle.Render(text)
		return lipgloss.JoinHorizontal(lipgloss.Top, cursor, styledText, countBadge)
	}

	styledText = MenuItemStyle.Render(text)
	return lipgloss.JoinHorizontal(lipgloss.Top, "  ", styledText, countBadge)
}

func RenderFileItem(filename string, isSelected bool, isBackOption bool) string {
	if isBackOption {
		if isSelected {
			cursor := CursorStyle.Render("❮ ")
			return cursor + SelectedBackOptionStyle.Render(filename)
		}
		return "  " + BackOptionStyle.Render(filename)
	}

	if isSelected {
		cursor := CursorStyle.Render("❯ ")
		return cursor + SelectedFileItemStyle.Render("📄 "+filename)
	}

	return "  " + FileItemStyle.Render("📄 "+filename)
}

func RenderSuccessMessage(message string) string {
	return SuccessStyle.Render("✅ " + message)
}

func RenderErrorMessage(message string) string {
	return ErrorStyle.Render("❌ " + message)
}

func RenderHelpText(text string) string {
	return HelpStyle.Render(text)
}
