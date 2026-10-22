// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/ring/pkg/ring/ringtest"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/xcli/pkg/xcli/xclitest"

	. "github.com/ctx42/xcli/pkg/xcli"
)

func Test_defaultFunc(t *testing.T) {
	// --- Given ---
	_, ccx := xclitest.New(t).CmdAndCtx("name", xclitest.NopFunc)

	// --- When ---
	err := DefaultFunc(nil, ccx)

	// --- Then ---
	assert.ErrorIs(t, ErrNoDefault, err)
	xrrtest.AssertEqual(t, "default command not defined for `name`", err)
}

func Test_Marshal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tt := []struct {
			testN string

			v   any
			exp string
		}{
			{"error", errors.New("msg"), "msg"},
			{"Marshaler", TMarshaler("test"), "test"},
			{
				"TextMarshaler",
				time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
				"2026-01-02T03:04:05Z",
			},
			{"Stringer", time.Second, "1s"},
			{"string", "abc", "abc"},
			{"bytes", []byte("abc"), "abc"},
			{"int", 42, "42"},
			{"nil", nil, "<nil>"},
		}

		for _, tc := range tt {
			t.Run(tc.testN, func(t *testing.T) {
				// --- When ---
				got, err := Marshal(tc.v)

				// --- Then ---
				assert.NoError(t, err)
				assert.Equal(t, tc.exp, string(got))
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		// --- Given ---
		v := TErrMarshaler("test")

		// --- When ---
		data, err := Marshal(v)

		// --- Then ---
		assert.ErrorEqual(t, "test", err)
		assert.Nil(t, data)
	})
}

func Test_RunCmd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		cmd := NewCommand("root", xclitest.NopFunc)
		assert.NoError(t, cmd.Parse())

		// --- When ---
		err := RunCmd(ctx, tst.Ring(), cmd)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("run returns error", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		fn := func(_ context.Context, _ *CmdCtx) error {
			return errors.New("returns error")
		}
		cmd := NewCommand("root", fn)
		assert.NoError(t, cmd.Parse())

		// --- When ---
		err := RunCmd(ctx, tst.Ring(), cmd)

		// --- Then ---
		assert.ErrorEqual(t, "returns error", err)
	})

	t.Run("run panics with string", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		fn := func(_ context.Context, _ *CmdCtx) error {
			panic("panics with string")
		}
		cmd := NewCommand("root", fn)
		assert.NoError(t, cmd.Parse())

		// --- When ---
		err := RunCmd(ctx, tst.Ring(), cmd)

		// --- Then ---
		assert.ErrorEqual(t, "panics with string", err)
	})

	t.Run("run panics with error", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		fn := func(_ context.Context, _ *CmdCtx) error {
			panic(errors.New("panics with error"))
		}
		cmd := NewCommand("root", fn)
		assert.NoError(t, cmd.Parse())

		// --- When ---
		err := RunCmd(ctx, tst.Ring(), cmd)

		// --- Then ---
		assert.ErrorEqual(t, "panics with error", err)
	})

	t.Run("run panics with Stringer", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		fn := func(_ context.Context, _ *CmdCtx) error {
			buf := &strings.Builder{}
			_, _ = buf.WriteString("panics with Stringer")
			panic(buf)
		}
		cmd := NewCommand("root", fn)
		assert.NoError(t, cmd.Parse())

		// --- When ---
		err := RunCmd(ctx, tst.Ring(), cmd)

		// --- Then ---
		assert.ErrorEqual(t, "panics with Stringer", err)
	})

	t.Run("run panics with unsupported type", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		fn := func(_ context.Context, _ *CmdCtx) error { panic(123) }
		cmd := NewCommand("root", fn)
		assert.NoError(t, cmd.Parse())

		// --- When ---
		err := RunCmd(ctx, tst.Ring(), cmd)

		// --- Then ---
		assert.ErrorEqual(t, "do not know how to handle `int` panics", err)
	})
}

func Test_DetectTTY_IsStdOutTTY_IsStdErrTTY(t *testing.T) {
	t.Run("stdout and stderr files", func(t *testing.T) {
		// --- Given ---
		sout := oskit.Open(t, "testdata/file0.txt")
		eout := oskit.Open(t, "testdata/file1.txt")

		tst := ringtest.New(t)
		rng := tst.Ring()
		rng.SetStdout(sout)
		rng.SetStderr(eout)

		// --- When ---
		have := DetectTTY(rng)

		// --- Then ---
		assert.False(t, IsStdOutTTY(have))
		assert.False(t, IsStdErrTTY(have))
	})

	t.Run("stdout and stderr not terminals", func(t *testing.T) {
		// --- Given ---
		tst := ringtest.New(t)
		rng := tst.Ring()

		// --- When ---
		have := DetectTTY(rng)

		// --- Then ---
		assert.False(t, IsStdOutTTY(have))
		assert.False(t, IsStdErrTTY(have))
	})

	t.Run("will not override values", func(t *testing.T) {
		// --- Given ---
		tst := ringtest.New(t)
		rng := tst.Ring()
		rng.MetaSet(MetaKeySOTerm, true)
		rng.MetaSet(MetaKeySETerm, true)

		// --- When ---
		have := DetectTTY(rng)

		// --- Then ---
		assert.True(t, IsStdOutTTY(have))
		assert.True(t, IsStdErrTTY(have))
	})

	t.Run("metadata is not set", func(t *testing.T) {
		// --- Given ---
		tst := ringtest.New(t)
		rng := tst.Ring()

		// --- Then ---
		assert.False(t, IsStdOutTTY(rng))
		assert.False(t, IsStdErrTTY(rng))
	})
}
