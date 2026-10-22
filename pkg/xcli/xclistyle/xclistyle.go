// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package xclistyle provides predefined lipgloss styles for command line
// program output.
package xclistyle

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

// Style represents predefined CLI styles.
type Style struct {
	Base    lipgloss.Style // Plain style the others derive from.
	Header  lipgloss.Style // Bold, near-white.
	Section lipgloss.Style // Bold, underlined, near-white.
	Error   lipgloss.Style // Bold red.
}

// New returns new CLI styles rendered for the given output.
func New(out io.Writer) Style {
	base := lipgloss.NewRenderer(out).NewStyle()
	sty := Style{Base: base}
	sty.Header = base.
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))
	sty.Section = base.Bold(true).
		Underline(true).
		Foreground(lipgloss.Color("#FAFAFA"))
	sty.Error = base.
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))
	return sty
}
