// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"flag"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/xflag/pkg/xflag"

	"github.com/ctx42/xcli/pkg/xcli"
)

func Test_WithEnv(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		// --- Given ---
		env := []string{"A=0", "B=1"}
		def := &options{}

		// --- When ---
		WithEnv(env)(def)

		// --- Then ---
		assert.Equal(t, []string{"A=0", "B=1"}, def.env)
		assert.NotSame(t, env, def.env)
	})

	t.Run("not aliased by WithEnvKey", func(t *testing.T) {
		// --- Given ---
		env := make([]string, 1, 2)
		env[0] = "A=0"
		def := &options{}
		WithEnv(env)(def)

		// --- When ---
		WithEnvKey("B", "1")(def)

		// --- Then ---
		assert.Equal(t, []string{"A=0", "B=1"}, def.env)
		assert.Equal(t, []string{"A=0", ""}, env[:2])
	})
}

func Test_WithEnvKey(t *testing.T) {
	// --- Given ---
	def := &options{env: []string{"A=0", "B=1"}}

	// --- When ---
	WithEnvKey("C", "2")(def)

	// --- Then ---
	assert.Equal(t, []string{"A=0", "B=1", "C=2"}, def.env)
}

func Test_WithTerm(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithTerm(true)(def)

	// --- Then ---
	assert.True(t, def.term)
}

func Test_WithClock(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithClock(time.Now)(def)

	// --- Then ---
	assert.Same(t, time.Now, def.clock)
}

func Test_New(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		// --- When ---
		tst := New(tspy)

		// --- Then ---
		rng := tst.Ring()
		assert.Equal(t, sorted(os.Environ()), sorted(rng.EnvAll()))
		assert.Equal(t, true, rng.MetaGet(xcli.MetaKeySOTerm))
		assert.Equal(t, true, rng.MetaGet(xcli.MetaKeySETerm))
		assert.Equal(t, false, rng.MetaGet(xcli.MetaKeyColor))
	})

	t.Run("with environment", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		// --- When ---
		tst := New(tspy, WithEnv([]string{"A=B", "C=D"}))

		// --- Then ---
		rng := tst.Ring()
		assert.Equal(t, []string{"A=B", "C=D"}, sorted(rng.EnvAll()))
	})

	t.Run("not on terminal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		// --- When ---
		tst := New(tspy, WithTerm(false))

		// --- Then ---
		rng := tst.Ring()
		assert.Equal(t, false, rng.MetaGet(xcli.MetaKeySOTerm))
		assert.Equal(t, false, rng.MetaGet(xcli.MetaKeySETerm))
		assert.Equal(t, false, rng.MetaGet(xcli.MetaKeyColor))
	})
}

func Test_Tester_WetStdout(t *testing.T) {
	t.Run("want stdout wet but is dry", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(3)
		tspy.ExpectError()
		tspy.ExpectLogEqual("expected buffer not to be empty:\n  name: stdout")
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.WetStdout()

		// --- Then ---
		assert.Same(t, tst, have)
	})

	t.Run("want stdout dry but is wet", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectError()
		want := "" +
			"expected buffer to be empty:\n" +
			"  name: stdout\n" +
			"  want: <empty>\n" +
			"  have: abc"
		tspy.ExpectLogEqual(want)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		_, _ = io.WriteString(tst.Streams().Stdout(), "abc")
	})
}

func Test_Tester_WetStderr(t *testing.T) {
	t.Run("want stderr wet but is dry", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(3)
		tspy.ExpectError()
		tspy.ExpectLogEqual("expected buffer not to be empty:\n  name: stderr")
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.WetStderr()

		// --- Then ---
		assert.Same(t, tst, have)
	})

	t.Run("want stderr dry but is wet", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectError()
		want := "" +
			"expected buffer to be empty:\n" +
			"  name: stderr\n" +
			"  want: <empty>\n" +
			"  have: abc"
		tspy.ExpectLogEqual(want)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		_, _ = io.WriteString(tst.Streams().Stderr(), "abc")
	})
}

func Test_Tester_Cmd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		cmd := tst.Cmd("name", DbgFunc, "a", "b", "c")

		// --- Then ---
		assert.Equal(t, "name", cmd.Name())
		assert.Equal(t, []string{"a", "b", "c"}, cmd.Args())
	})

	t.Run("error - unknown flag", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectError()
		tspy.ExpectLogEqual("flag provided but not defined: -abc")
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		cmd := tst.Cmd("name", NopFunc, "--abc", "b")

		// --- Then ---
		assert.Nil(t, cmd)
	})
}

func Test_Tester_Ctx(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		tst := New(tspy)
		fs := xflag.NewFlagSet("name", flag.ContinueOnError)
		fs.Bool("abc", false, "usage")
		cmd := xcli.NewCommand("name", nil, xcli.WithFlags(fs))

		// --- When ---
		ccx := tst.Ctx(cmd)

		// --- Then ---
		assert.Same(t, cmd, ccx.Cmd)
		assert.True(t, cmd.Parsed())
		assert.Equal(t, tst.Ring(), ccx.Rng)
		assert.False(t, cmd.GetBool("name"))
	})

	t.Run("with args", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		tst := New(tspy)
		fs := xflag.NewFlagSet("name", flag.ContinueOnError)
		fs.Bool("abc", false, "usage")
		cmd := xcli.NewCommand("name", nil, xcli.WithFlags(fs))

		// --- When ---
		ccx := tst.Ctx(cmd, "--abc")

		// --- Then ---
		assert.Same(t, cmd, ccx.Cmd)
		assert.True(t, cmd.Parsed())
		assert.Equal(t, tst.Ring(), ccx.Rng)
		assert.True(t, cmd.GetBool("abc"))
	})

	t.Run("error - parse", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectError()
		tspy.ExpectLogEqual("flag provided but not defined: -unknown")
		tspy.Close()

		tst := New(tspy)
		fs := xflag.NewFlagSet("name", flag.ContinueOnError)
		fs.Bool("abc", false, "usage")
		cmd := xcli.NewCommand("name", nil, xcli.WithFlags(fs))

		// --- When ---
		ccx := tst.Ctx(cmd, "--unknown")

		// --- Then ---
		assert.Nil(t, ccx)
		assert.True(t, cmd.Parsed())
		assert.False(t, cmd.GetBool("name"))
	})
}

func Test_Tester_CmdAndCtx(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		cmd, ccx := tst.CmdAndCtx("name", NopFunc, "a", "b", "c")

		// --- Then ---
		assert.Same(t, cmd, ccx.Cmd)
		assert.Equal(t, "name", ccx.Cmd.Name())
		assert.Equal(t, []string{"a", "b", "c"}, ccx.Cmd.Args())
		assert.Same(t, tst.Streams().Stdin(), ccx.Rng.Stdin())
		assert.Same(t, tst.Streams().Stdout(), ccx.Rng.Stdout())
		assert.Same(t, tst.Streams().Stderr(), ccx.Rng.Stderr())
		assert.Equal(t, sorted(tst.Ring().EnvAll()), sorted(ccx.Rng.EnvAll()))
		assert.NotSame(t, tst.Ring().EnvAll(), ccx.Rng.EnvAll())
	})

	t.Run("error - parse", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectError()
		tspy.ExpectLogEqual("flag provided but not defined: -abc")
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		cmd, ccx := tst.CmdAndCtx("name", NopFunc, "--abc")

		// --- Then ---
		assert.Nil(t, cmd)
		assert.Nil(t, ccx)
	})
}

func Test_Tester_Func(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(3)
	tspy.Close()

	tst := New(tspy).WetStdout()
	_, cmx := tst.CmdAndCtx("name", NopFunc, "a", "b", "c")

	// --- When ---
	fn := tst.Func()

	// --- Then ---
	assert.NotNil(t, fn)

	assert.NoError(t, fn(t.Context(), cmx))
	assert.Equal(t, "command `name` with args [a b c]", tst.Stdout())
}

func Test_Tester_NopFunc(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(2)
	tspy.Close()

	tst := New(tspy)
	_, cmx := tst.CmdAndCtx("name", NopFunc, "a", "b", "c")

	// --- When ---
	fn := tst.NopFunc()

	// --- Then ---
	assert.NotNil(t, fn)

	assert.NoError(t, fn(t.Context(), cmx))
	assert.Equal(t, "", tst.Stderr())
}
