// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"bytes"
	"testing"

	"github.com/ctx42/ring/pkg/ring"
	"github.com/ctx42/testing/pkg/assert"

	"github.com/ctx42/xcli/pkg/xcli/internal/fixture"

	. "github.com/ctx42/xcli/pkg/xcli"
)

func Test_NewCmdCtx(t *testing.T) {
	// --- Given ---
	buf := &bytes.Buffer{}
	rng := ring.New()
	rng.SetStdout(buf)
	cmd := fixture.TstCmd()

	// --- When ---
	ccx := NewCmdCtx(rng, cmd)

	// --- Then ---
	assert.Same(t, buf, ccx.Rng.Stdout())
	assert.Same(t, cmd, ccx.Cmd)
}
