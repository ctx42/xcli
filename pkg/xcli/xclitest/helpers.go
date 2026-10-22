// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/ctx42/xcli/pkg/xcli"
)

// DbgFunc is a test [xcli.Func] which prints debug info to stdout.
func DbgFunc(_ context.Context, ccx *xcli.CmdCtx) error {
	const format = "command %#q with args %v"
	rng := ccx.Rng
	cmd := ccx.Cmd
	_, err := fmt.Fprintf(rng.Stdout(), format, cmd.Name(), cmd.Args())
	return err
}

// NopFunc is a no operation [xcli.Func].
func NopFunc(_ context.Context, _ *xcli.CmdCtx) error { return nil }

// TstCmd returns test command with sub commands.
//
//	        ┌────────┐
//	        │  ROOT  │
//	        └────┬───┘
//	     ┌───────┴───────┐
//	┌────┴────┐    ┌─────┴────┐
//	│  sub-a  │    │  sub-b   │
//	└─────────┘    └────┬─────┘
//	               ┌────┴─────┐
//	               │  sub-b-a │
//	               └──────────┘
func TstCmd() *xcli.Command {
	// Defines usage flag and its alias.
	helpFlg := func(cmd *xcli.Command, usage string) {
		cmd.BoolSL("help", "h", false, usage)
	}

	subA := xcli.NewCommand("sub-a", DbgFunc).
		ShortDesc("subcommand A").
		ArgsRule(xcli.ArgsRule(0))
	subA.Int("sa", 42, "sa-help")

	subBA := xcli.NewCommand("sub-b-a", DbgFunc).
		ShortDesc("subcommand BA")
	subBA.String("sba", "sba-def", "sba-help")
	helpFlg(subBA, "sub-b-help")

	subB := xcli.NewCommand("sub-b", DbgFunc).
		ShortDesc("subcommand B").
		ArgsRule(xcli.MinArgsRule(2)).
		UsageFlagName("help").
		Add(subBA)
	helpFlg(subB, "sub-b-help")
	subB.String("sb", "sb-def", "sb-help")

	root := xcli.NewCommand("root", DbgFunc).
		ShortDesc("main command").
		ArgsRule(xcli.ArgsRule(0)).
		Example("root --root-def abc", "using long form flag").
		Example("root -r abc", "").
		Example("root -r abc", "using short form flag").
		UsageFlagName("help").
		Add(subA, subB)
	helpFlg(root, "root-help")
	root.StringSL("root", "r", "root-def", "root-help")

	return root
}

// TstCmdRequired returns test command with sub commands and one of them has
// required flag.
//
//	        ┌────────┐
//	        │  ROOT  │
//	        └────┬───┘
//	     ┌───────┴───────┐
//	┌────┴────┐    ┌─────┴────┐
//	│  sub-a  │    │  sub-b*  │
//	└─────────┘    └────┬─────┘
//	               ┌────┴─────┐
//	               │  sub-b-a │
//	               └──────────┘
//
// The ones marked with a `*` have required flag.
func TstCmdRequired() *xcli.Command {
	// Defines usage flag and its alias.
	helpFlg := func(cmd *xcli.Command, usage string) {
		cmd.BoolSL("help", "h", false, usage)
	}

	subA := xcli.NewCommand("sub-a", DbgFunc).
		ShortDesc("subcommand A").
		UsageFlagName("help").
		ArgsRule(xcli.ArgsRule(0))
	helpFlg(subA, "sub-a-help")

	subBA := xcli.NewCommand("sub-b-a", DbgFunc).
		ShortDesc("subcommand BA").
		ArgsRule(xcli.ArgsRule(0)).
		UsageFlagName("help")
	subBA.String("sba", "sba-def", "sba-help")
	helpFlg(subBA, "sub-b-a-help")

	subB := xcli.NewCommand("sub-b", DbgFunc).
		ShortDesc("subcommand B").
		ArgsRule(xcli.ArgsRule(0)).
		UsageFlagName("help").
		Add(subBA)
	helpFlg(subB, "sub-b-help")
	subB.String("sb", "", "sb-help")
	subB.Required("sb")

	root := xcli.NewCommand("root", DbgFunc).
		Add(subA, subB)

	return root
}

// GenPre returns function matching [xcli.PreRun] signature which when called
// sets the key to the value in the [xcli.Command] metadata.
func GenPre(key string, value int) xcli.PreRun {
	return func(
		ctx context.Context,
		ccx *xcli.CmdCtx,
	) (context.Context, error) {

		ccx.Cmd.MetaSet(key, value)
		return ctx, nil
	}
}

// TstCmdPreRun returns test command designed to test pre run functionality.
//
//	         ┌───────┐
//	         │ ROOT* │
//	         └───┬───┘
//	     ┌───────┴─────────┐
//	┌────┴──────┐    ┌─────┴─────┐
//	│  sub-a    │    │  sub-b*   │
//	└────┬──────┘    └─────┬─────┘
//	┌────┴──────┐    ┌─────┴─────┐
//	│  sub-a-a* │    │  sub-b-a* │
//	└───────────┘    └───────────┘
//
// The ones marked with a `*` have pre runners declared.
func TstCmdPreRun() *xcli.Command {
	fn := func(_ context.Context, ccx *xcli.CmdCtx) error {
		eout := ccx.Rng.Stderr()
		meta := ccx.Cmd.MetaAll()
		keys := slices.Sorted(maps.Keys(meta))
		_, _ = fmt.Fprintf(eout, "dbg-pre:")
		for _, key := range keys {
			_, _ = fmt.Fprintf(eout, " %s: %v\n", key, meta[key])
		}
		return nil
	}

	subBA := xcli.NewCommand("sub-b-a", fn).
		ArgsRule(xcli.ArgsRule(0)).
		PreRun(GenPre("sub-b-a", 1))

	subAA := xcli.NewCommand("sub-a-a", fn).
		ArgsRule(xcli.ArgsRule(0)).
		PreRun(GenPre("sub-a-a", 1))

	subA := xcli.NewCommand("sub-a", fn).
		ArgsRule(xcli.ArgsRule(0)).
		Add(subAA)

	subB := xcli.NewCommand("sub-b", fn).
		ShortDesc("subcommand B").
		ArgsRule(xcli.ArgsRule(0)).
		PreRun(GenPre("sub-b", 1)).
		Add(subBA)

	root := xcli.NewCommand("root", fn).
		ArgsRule(xcli.ArgsRule(0)).
		PreRun(GenPre("root", 1)).
		Add(subA, subB)

	return root
}
