// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"context"
	"flag"
	"testing"

	"github.com/ctx42/ring/pkg/ring/ringtest"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xflag/pkg/xflag"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/xcli/pkg/xcli/xclitest"

	. "github.com/ctx42/xcli/pkg/xcli"
)

func Test_WithFlags(t *testing.T) {
	// --- Given ---
	fs := xflag.NewFlagSet("name", flag.ContinueOnError)
	cmd := &Command{}

	// --- When ---
	WithFlags(fs)(cmd)

	// --- Then ---
	assert.Equal(t, "name", cmd.Name())
}

func Test_NewCommand(t *testing.T) {
	t.Run("with Function", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := xclitest.New(t).WetStdout()

		// --- When ---
		cmd := NewCommand("root", tst.Func())

		// --- Then ---
		assert.NoError(t, cmd.Parse())
		assert.NoError(t, cmd.Run(ctx, tst.Ring()))
		assert.Equal(t, "command `root` with args []", tst.Stdout())
	})

	t.Run("command without function", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := xclitest.New(t)

		// --- When ---
		cmd := NewCommand("root", nil)

		// --- Then ---
		assert.NoError(t, cmd.Parse())

		err := cmd.Run(ctx, tst.Ring())
		assert.ErrorIs(t, ErrNoDefault, err)
		xrrtest.AssertEqual(t, "default command not defined for `root`", err)
	})

	t.Run("command requires subcommand to be defined", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := xclitest.New(t)

		// --- When ---
		root := NewCommand("root", nil)
		root.Add(NewCommand("sub", xclitest.NopFunc))

		// --- Then ---
		assert.NoError(t, root.Parse())

		err := root.Run(ctx, tst.Ring())
		assert.ErrorIs(t, ErrSubCmdReq, err)
		xrrtest.AssertEqual(t, "subcommand required for `root`", err)
	})

	t.Run("command name overrides flag set name", func(t *testing.T) {
		// --- Given ---
		fs := xflag.NewFlagSet("name", flag.ContinueOnError)

		// --- When ---
		cmd := NewCommand("root", nil, WithFlags(fs))

		// --- Then ---
		assert.Equal(t, "root", cmd.Name())
	})
}

func Test_Command_Add(t *testing.T) {
	t.Run("add multiple", func(t *testing.T) {
		// --- Given ---
		root := NewCommand("root", nil)
		sa := NewCommand("sub-a", nil)
		sba := NewCommand("sub-b-a", nil)
		sb := NewCommand("sub-b", nil).Add(sba)

		// --- When ---
		have := root.Add(sa, sb)

		// --- Then ---
		assert.Same(t, root, have)
		wantDesc := "" +
			"cmd: root\n" +
			"  cmd: sub-a\n" +
			"  cmd: sub-b\n" +
			"    cmd: sub-b-a\n"
		assert.Equal(t, wantDesc, have.Describe())
		assert.Same(t, root.MetaAll(), sa.MetaAll())
		assert.Same(t, root.MetaAll(), sb.MetaAll())
		assert.Same(t, root.MetaAll(), sba.MetaAll())
	})
}

func Test_Command_UsageFlagName(t *testing.T) {
	t.Run("add bool", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("root", nil)
		cmd.Bool("help", false, "show help")

		// --- When ---
		have := cmd.UsageFlagName("help")

		// --- Then ---
		assert.Same(t, cmd, have)
		assert.NoError(t, cmd.Parse("--help"))
		assert.True(t, cmd.GetBool("help"))
	})
}

func Test_Command_ShouldRunUsage(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("root", nil).UsageFlagName("help")
		cmd.Bool("help", false, "show help")
		assert.NoError(t, cmd.Parse("--help"))

		// --- When ---
		have := cmd.ShouldRunUsage()

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("root", nil).UsageFlagName("help")
		cmd.Bool("help", false, "show help")

		// --- When ---
		have := cmd.ShouldRunUsage()

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("always true", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("root", nil).UsageFlagName("help")
		cmd.Bool("help", true, "show help")

		// --- When ---
		have := cmd.ShouldRunUsage()

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true when parent command help requested", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStderr()

		sub := NewCommand("sub", xclitest.DbgFunc).
			UsageFlagName("help")
		sub.Bool("help", false, "sub-desc")

		root := NewCommand("root", nil).
			UsageFlagName("help").
			Add(sub)
		root.Bool("help", false, "root-desc")
		assert.NoError(t, root.Parse("--help", "sub"))

		// --- When ---
		have := root.ShouldRunUsage()

		// --- Then ---
		assert.True(t, have)
		assert.NoError(t, root.Run(ctx, tst.Ring()))

		want := "" +
			"Usage of `root sub`:\n" +
			"                \n" +
			"OPTIONS:        \n" +
			"      --help    sub-desc\n"
		assert.Equal(t, want, tst.Stderr())
		assert.Empty(t, root.Args())
	})
}

func Test_Command_ArgsRule(t *testing.T) {
	// --- Given ---
	tst := xclitest.New(t)
	cmd := tst.Cmd("root", xclitest.NopFunc, "arg0", "arg1")
	cmd.ArgsRule(MinArgsRule(3))

	// --- When ---
	err := cmd.Run(context.Background(), tst.Ring())

	// --- Then ---
	wMsg := "command `root` requires minimum 3 arguments 2 given"
	assert.ErrorEqual(t, wMsg, err)
	assert.Equal(t, []string{"arg0", "arg1"}, cmd.Args())
}

func Test_Command_Find(t *testing.T) {
	t.Run("find leaf command", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()

		// --- When ---
		have := root.Find("root", "sub-b", "sub-b-a")

		// --- Then ---
		assert.NotNil(t, have)
		assert.Equal(t, "sub-b-a", have.Name())
	})

	t.Run("find first level command", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()

		// --- When ---
		have := root.Find("root", "sub-b")

		// --- Then ---
		assert.NotNil(t, have)
		assert.Equal(t, "sub-b", have.Name())
	})

	t.Run("find root command", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()

		// --- When ---
		have := root.Find("root")

		// --- Then ---
		assert.NotNil(t, have)
		assert.Equal(t, "root", have.Name())
	})

	t.Run("unknown leaf command", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()

		// --- When ---
		have := root.Find("root", "sub-b", "sub-b-a", "sub-b-a-a")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("unknown command", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()

		// --- When ---
		have := root.Find("unknown")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_Parse(t *testing.T) {
	t.Run("parse flags", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd()
		cmd.String("str", "abc", "usage")

		// --- When ---
		err := cmd.Parse("--root", "abc", "--str", "xyz")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "abc", cmd.GetFlagSet("root").GetString("root"))
		assert.Equal(t, "xyz", cmd.GetFlagSet("root").GetString("str"))
	})

	t.Run("first level subcommand flag and arguments", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()
		args := []string{"sub-b", "--sb", "abc", "arg0", "arg1"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		assert.NoError(t, err)
		wantDesc := "" +
			"cmd(sub-b): root --help false --root root-def\n" +
			"  cmd: sub-a --sa 42\n" +
			"  cmd: sub-b --help false --sb abc\n" +
			"    cmd: sub-b-a --help false --sba sba-def\n"
		assert.Equal(t, wantDesc, root.Describe())
		assert.Equal(t, []string{"arg0", "arg1"}, root.Args())
	})

	t.Run("unknown flag", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd()

		// --- When ---
		err := cmd.Parse("--unknown")

		// --- Then ---
		assert.ErrorEqual(t, "flag provided but not defined: -unknown", err)
	})

	t.Run("optional string flag", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd()
		cmd.String("str", "abc", "usage")

		// --- When ---
		err := cmd.Parse()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "abc", cmd.GetFlagSet("root").GetString("str"))
	})

	t.Run("subcommand parsing error", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()
		args := []string{"sub-a", "--sa", "abc"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		wMsg := `invalid value "abc" for flag -sa: parse error`
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("required flag must be provided", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmdRequired()
		args := []string{"sub-b", "sub-b-a"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		assert.ErrorEqual(t, "`sb` flag is required for command `sub-b`", err)
	})

	t.Run("root help requested", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()
		args := []string{"--help"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("subcommand help requested", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()
		args := []string{"sub-b", "--help"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("sub subcommand help requested", func(t *testing.T) {
		// --- Given ---
		root := xclitest.TstCmd()
		args := []string{"sub-b", "sub-b-a", "--help"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("sub subcommand help requested", func(t *testing.T) {
		// --- Given ---
		subA := NewCommand("sub-a", xclitest.DbgFunc)
		root := NewCommand("root", xclitest.DbgFunc).Add(subA)
		args := []string{"sub-a", "--help"}

		// --- When ---
		err := root.Parse(args...)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_Command_Run(t *testing.T) {
	t.Run("run root without arguments", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStdout()
		root := xclitest.TstCmd()
		assert.NoError(t, root.Parse())

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "command `root` with args []", tst.Stdout())
		wantDesc := "" +
			"cmd(root): root --help false --root root-def\n" +
			"  cmd: sub-a --sa 42\n" +
			"  cmd: sub-b --help false --sb sb-def\n" +
			"    cmd: sub-b-a --help false --sba sba-def\n"
		assert.Equal(t, wantDesc, root.Describe())
	})

	t.Run("run root with flag", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStdout()
		root := xclitest.TstCmd()
		assert.NoError(t, root.Parse("--root", "abc"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "command `root` with args []", tst.Stdout())
		wantDesc := "" +
			"cmd(root): root --help false --root abc\n" +
			"  cmd: sub-a --sa 42\n" +
			"  cmd: sub-b --help false --sb sb-def\n" +
			"    cmd: sub-b-a --help false --sba sba-def\n"
		assert.Equal(t, wantDesc, root.Describe())
		assert.Empty(t, root.Args())
	})

	t.Run("run first level subcommand flag", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStdout()
		root := xclitest.TstCmd()
		assert.NoError(t, root.Parse("sub-b", "--sb", "abc", "arg0", "arg1"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "command `sub-b` with args [arg0 arg1]", tst.Stdout())
		wantDesc := "" +
			"cmd(sub-b): root --help false --root root-def\n" +
			"  cmd: sub-a --sa 42\n" +
			"  cmd: sub-b --help false --sb abc\n" +
			"    cmd: sub-b-a --help false --sba sba-def\n"
		assert.Equal(t, wantDesc, root.Describe())
		assert.Equal(t, []string{"arg0", "arg1"}, root.Args())
	})

	t.Run("run first level subcommand flag and arguments", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStdout()
		root := xclitest.TstCmd()
		args := []string{"sub-b", "--sb", "abc", "arg0", "arg1"}
		assert.NoError(t, root.Parse(args...))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "command `sub-b` with args [arg0 arg1]", tst.Stdout())
		wantDesc := "" +
			"cmd(sub-b): root --help false --root root-def\n" +
			"  cmd: sub-a --sa 42\n" +
			"  cmd: sub-b --help false --sb abc\n" +
			"    cmd: sub-b-a --help false --sba sba-def\n"
		assert.Equal(t, wantDesc, root.Describe())
		assert.Equal(t, []string{"arg0", "arg1"}, root.Args())
	})

	t.Run("run usage", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		root := xclitest.TstCmd()
		assert.NoError(t, root.Parse("--help"))

		var usageRan bool
		root.Usage = func() { usageRan = true }

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		assert.True(t, usageRan)
		assert.Empty(t, root.Args())
	})

	t.Run("show help for subcommand", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStderr()
		root := xclitest.TstCmd()
		assert.NoError(t, root.Parse("sub-b", "-h"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		want := "" +
			"Usage of `root sub-b`:\n\n" +
			"  subcommand B\n" +
			"                \n" +
			"OPTIONS:        \n" +
			"  -h, --help    sub-b-help\n" +
			"      --sb      sb-help\n" +
			"                \n" +
			"COMMANDS:       \n" +
			"  sub-b-a       subcommand BA\n"
		assert.Equal(t, want, tst.Stderr())
		assert.Empty(t, root.Args())
	})

	t.Run("show help requested before subcommand with required flag", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStderr()
		root := xclitest.TstCmdRequired()
		assert.NoError(t, root.Parse("sub-b", "-h", "sub-b-a"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		want := "" +
			"Usage of `root sub-b sub-b-a`:\n\n" +
			"  subcommand BA\n" +
			"                \n" +
			"OPTIONS:        \n" +
			"  -h, --help    sub-a-help\n" +
			"      --sba     sba-help\n"
		assert.Equal(t, want, tst.Stderr())
		assert.Empty(t, root.Args())
	})

	t.Run("arguments count error", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		root := xclitest.TstCmd()
		assert.NoError(t, root.Parse("sub-a", "--sa", "44", "arg0"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		wMsg := "command `sub-a` requires no arguments 1 given"
		assert.ErrorEqual(t, wMsg, err)
		assert.Equal(t, []string{"arg0"}, root.Args())
	})

	t.Run("Parse method must be executed before Run", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t)
		root := xclitest.TstCmd()

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		wMsg := "not parsed arguments"
		assert.ErrorEqual(t, wMsg, err)
	})
}

func Test_Command_PreRun_RunPreRun(t *testing.T) {
	t.Run("pre run error", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		pre := func(_ context.Context, _ *CmdCtx) (context.Context, error) {
			return nil, ErrTest
		}
		cmd, ccx := xclitest.New(t).CmdAndCtx("root", nil)

		// --- When ---
		have := cmd.PreRun(pre)

		// --- Then ---
		assert.Same(t, cmd, have)

		// --- When ---
		haveCtx, err := RunPreRun(ctx, ccx)

		// --- Then ---
		assert.Nil(t, haveCtx)
		assert.ErrorIs(t, ErrTest, err)
	})

	t.Run("stops at first error", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()

		var pre0ran, pre1ran bool
		pre0 := func(_ context.Context, _ *CmdCtx) (context.Context, error) {
			pre0ran = true
			return nil, ErrTest
		}

		pre1 := func(ctx context.Context, ccx *CmdCtx) (context.Context, error) {
			pre1ran = true
			return ctx, nil
		}

		cmd, ccx := xclitest.New(t).CmdAndCtx("root", nil)

		// --- When ---
		have := cmd.PreRun(pre0, pre1)

		// --- Then ---
		assert.Same(t, cmd, have)

		// --- When ---
		haveCtx, err := RunPreRun(ctx, ccx)

		// --- Then ---
		assert.Nil(t, haveCtx)
		assert.ErrorIs(t, ErrTest, err)
		assert.True(t, pre0ran)
		assert.False(t, pre1ran)
	})
}

func Test_Command_PreRun_Run_RunPreRun(t *testing.T) {
	t.Run("all commands in the chain have run pre-run", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStderr()
		root := xclitest.TstCmdPreRun()
		root.MetaSet("my", 1)
		assert.NoError(t, root.Parse("sub-b", "sub-b-a"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		want := "dbg-pre: my: 1\n root: 1\n sub-b: 1\n sub-b-a: 1\n"
		assert.Equal(t, want, tst.Stderr())
		assert.Empty(t, root.Args())
	})

	t.Run("some commands in the chain have run pre-run", func(t *testing.T) {
		// --- Given ---
		ctx := context.Background()
		tst := ringtest.New(t).WetStderr()
		root := xclitest.TstCmdPreRun()
		root.MetaSet("my", 1)
		assert.NoError(t, root.Parse("sub-a", "sub-a-a"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.NoError(t, err)
		want := "dbg-pre: my: 1\n root: 1\n sub-a-a: 1\n"
		assert.Equal(t, want, tst.Stderr())
		assert.Empty(t, root.Args())
	})

	t.Run("run pre-run error", func(t *testing.T) {
		// --- Given ---
		pre := func(_ context.Context, _ *CmdCtx) (context.Context, error) {
			return nil, ErrTest
		}

		ctx := context.Background()
		tst := ringtest.New(t)
		root := xclitest.TstCmdPreRun().PreRun(pre)
		root.MetaSet("my", 1)
		assert.NoError(t, root.Parse("sub-a", "sub-a-a"))

		// --- When ---
		err := root.Run(ctx, tst.Ring())

		// --- Then ---
		assert.ErrorIs(t, ErrTest, err)
		assert.Empty(t, root.Args())
	})
}

func Test_Command_GetFlagSet(t *testing.T) {
	subA := NewCommand("sub-a", xclitest.DbgFunc)
	subA.Int("sa", 42, "sa-desc")

	subB := NewCommand("sub-b", xclitest.DbgFunc)
	subB.String("sb", "sb-def", "sb-desc")

	subBA := NewCommand("sub-b-a", xclitest.DbgFunc)
	subBA.String("sba", "sba-def", "sba-desc")
	subB.Add(subBA)

	root := NewCommand("root", xclitest.DbgFunc).Add(subA, subB)
	root.String("root", "root-def", "root-desc")

	t.Run("get default value of existing", func(t *testing.T) {
		// --- When ---
		have := subBA.GetFlagSet("root")

		// --- Then ---
		assert.Equal(t, "root-def", have.GetString("root"))
	})

	t.Run("unknown root", func(t *testing.T) {
		// --- When ---
		have := subBA.GetFlagSet("unknown")

		// --- Then ---
		assert.Empty(t, have.GetString("root"))
		assert.Equal(t, "unknown", have.Name())
	})

	t.Run("unknown root flag", func(t *testing.T) {
		// --- When ---
		have := subBA.GetFlagSet("root")

		// --- Then ---
		assert.Empty(t, have.GetString("unknown"))
		assert.Equal(t, "root", have.Name())
	})
}

func Test_Command_IsRoot(t *testing.T) {
	t.Run("is root", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("abc", nil)

		// --- When ---
		have := cmd.IsRoot()

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("is not root", func(t *testing.T) {
		// --- Given ---
		cmd0 := NewCommand("abc", nil)
		cmd1 := NewCommand("def", nil)
		cmd0.Add(cmd1)

		// --- When ---
		have := cmd1.IsRoot()

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Command_FullName(t *testing.T) {
	t.Run("root command", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd()

		// --- When ---
		have := cmd.FullName()

		// --- Then ---
		assert.Equal(t, "root", have)
	})

	t.Run("subcommand", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd().Find("sub-b", "sub-b-a")

		// --- When ---
		have := cmd.FullName()

		// --- Then ---
		assert.Equal(t, "root sub-b sub-b-a", have)
	})
}

func Test_Command_Help(t *testing.T) {
	t.Run("help for root", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd()

		// --- When ---
		have := cmd.Help()

		// --- Then ---
		want := "" +
			"Usage of `root`:\n\n" +
			"  main command\n\n" +
			"EXAMPLES:\n" +
			"  root --root-def abc    # using long form flag\n" +
			"  root -r abc            \n" +
			"  root -r abc            # using short form flag\n" +
			"                \n" +
			"OPTIONS:        \n" +
			"  -h, --help    root-help\n" +
			"  -r, --root    root-help\n" +
			"                \n" +
			"COMMANDS:       \n" +
			"  sub-a         subcommand A\n" +
			"  sub-b         subcommand B\n"
		assert.Equal(t, want, have)
	})

	t.Run("help for subcommand", func(t *testing.T) {
		// --- Given ---
		cmd := xclitest.TstCmd().Find("root", "sub-b")

		// --- When ---
		have := cmd.Help()

		// --- Then ---
		want := "" +
			"Usage of `root sub-b`:\n\n" +
			"  subcommand B\n" +
			"                \n" +
			"OPTIONS:        \n" +
			"  -h, --help    sub-b-help\n" +
			"      --sb      sb-help\n" +
			"                \n" +
			"COMMANDS:       \n" +
			"  sub-b-a       subcommand BA\n"
		assert.Equal(t, want, have)
	})

	t.Run("without short description", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("cmd", nil)

		// --- When ---
		have := cmd.Help()

		// --- Then ---
		want := "Usage of `cmd`:\n"
		assert.Equal(t, want, have)
	})

	t.Run("with short description", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("cmd", nil).ShortDesc("short desc")

		// --- When ---
		have := cmd.Help()

		// --- Then ---
		want := "Usage of `cmd`:\n\n  short desc\n"
		assert.Equal(t, want, have)
	})

	t.Run("with example", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("cmd", nil).Example("cmd", "do it")

		// --- When ---
		have := cmd.Help()

		// --- Then ---
		want := "Usage of `cmd`:\n\nEXAMPLES:\n  cmd    # do it\n"
		assert.Equal(t, want, have)
	})

	t.Run("with short and long description", func(t *testing.T) {
		// --- Given ---
		cmd := NewCommand("cmd", nil).
			ShortDesc("short desc").
			LongDesc("long desc")

		// --- When ---
		have := cmd.Help()

		// --- Then ---
		want := "Usage of `cmd`:\n\n  long desc\n"
		assert.Equal(t, want, have)
	})
}
