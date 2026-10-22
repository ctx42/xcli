// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctx42/testing/pkg/assert"

	. "github.com/ctx42/xcli/pkg/xcli"
)

func Test_NewStyle(t *testing.T) {
	// --- When ---
	have := NewStyle(&bytes.Buffer{})

	// --- Then ---
	assert.False(t, have.Base.GetBold())
	assert.Equal(t, lipgloss.NoColor{}, have.Base.GetForeground())

	assert.True(t, have.Header.GetBold())
	assert.False(t, have.Header.GetUnderline())
	assert.Equal(t, lipgloss.Color("#FAFAFA"), have.Header.GetForeground())

	assert.True(t, have.Section.GetBold())
	assert.True(t, have.Section.GetUnderline())
	assert.Equal(t, lipgloss.Color("#FAFAFA"), have.Section.GetForeground())

	assert.True(t, have.Error.GetBold())
	assert.False(t, have.Error.GetUnderline())
	assert.Equal(t, lipgloss.Color("#FF0000"), have.Error.GetForeground())
}
