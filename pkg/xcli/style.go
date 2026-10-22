// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

// Style represents predefined CLI styles.
type Style struct {
	Base    lipgloss.Style // Represents base style.
	Header  lipgloss.Style // Represents a CLI style for headers.
	Section lipgloss.Style // Represents a CLI style for section headers.
	Error   lipgloss.Style // Represents CLI style for error message.
}

// NewStyle returns new CLI styles based on output buffer.
func NewStyle(out io.Writer) Style {
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
