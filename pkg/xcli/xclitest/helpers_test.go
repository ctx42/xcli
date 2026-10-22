// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"context"
	"testing"

	"github.com/ctx42/testing/pkg/assert"

	"github.com/ctx42/xcli/pkg/xcli"
)

func Test_DbgFunc(t *testing.T) {
	// --- Given ---
	var called bool
	fn := func(ctx context.Context, ccx *xcli.CmdCtx) error {
		called = true
		return nil
	}
	tst := New(t).WetStdout()
	_, ccx := tst.CmdAndCtx("name", fn, "arg0", "arg1")

	// --- When ---
	err := DbgFunc(t.Context(), ccx)

	// --- Then ---
	assert.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "command `name` with args [arg0 arg1]", tst.Stdout())
}

func Test_NopFunc(t *testing.T) {
	// --- Given ---
	var err error

	// --- When ---
	assert.NoPanic(t, func() { err = NopFunc(t.Context(), &xcli.CmdCtx{}) })

	// --- Then ---
	assert.NoError(t, err)
}
