// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/ctx42/ring/pkg/ring"

	"github.com/ctx42/xcli/pkg/xcli"
)

func ExampleCommand() {
	greet := func(_ context.Context, ccx *xcli.CmdCtx) error {
		name := ccx.Cmd.GetString("name")
		_, err := fmt.Fprintf(ccx.Rng.Stdout(), "Hello, %s!\n", name)
		return err
	}

	cmd := xcli.NewCommand("hello", greet).
		ShortDesc("print a greeting").
		UsageFlagName("help")
	cmd.BoolSL("help", "h", false, "show help")
	cmd.StringSL("name", "n", "World", "who to greet")

	// A program passes os.Args[1:] instead.
	if err := cmd.Parse("--name", "gopher"); err != nil {
		fmt.Println(err)
		return
	}
	if err := xcli.RunCmd(context.Background(), ring.New(), cmd); err != nil {
		fmt.Println(err)
	}
	// Output:
	// Hello, gopher!
}

func ExampleCommand_Help() {
	cmd := xcli.NewCommand("hello", nil).
		ShortDesc("print a greeting").
		UsageFlagName("help")
	cmd.BoolSL("help", "h", false, "show help")
	cmd.StringSL("name", "n", "World", "who to greet")

	// Run prints the same text to standard error on -h or --help.
	fmt.Print(cmd.Help())
	// Output:
	// Usage of `hello`:
	//
	//   print a greeting
	//
	// OPTIONS:
	//   -h, --help    show help
	//   -n, --name    who to greet
}

func ExampleCommand_Add() {
	addCmd := xcli.NewCommand("add", nil).
		ShortDesc("add a task").
		ArgsRule(xcli.MinArgsRule(1))
	listCmd := xcli.NewCommand("list", nil).
		ShortDesc("list tasks")

	root := xcli.NewCommand("todo", nil).
		ShortDesc("manage a task list").
		Example("todo add buy milk", "add a task").
		Example("todo -f work.txt list", "list tasks in work.txt").
		UsageFlagName("help").
		Add(addCmd, listCmd)
	root.BoolSL("help", "h", false, "show help")
	root.StringSL("file", "f", "todo.txt", "task list file")

	fmt.Print(root.Help())
	// Output:
	// Usage of `todo`:
	//
	//   manage a task list
	//
	// EXAMPLES:
	//   todo add buy milk        # add a task
	//   todo -f work.txt list    # list tasks in work.txt
	//
	// OPTIONS:
	//   -f, --file    task list file
	//   -h, --help    show help
	//
	// COMMANDS:
	//   add           add a task
	//   list          list tasks
}

func ExampleCommand_Run_errors() {
	nop := func(context.Context, *xcli.CmdCtx) error { return nil }
	newRoot := func() *xcli.Command {
		addCmd := xcli.NewCommand("add", nop).ArgsRule(xcli.MinArgsRule(1))
		listCmd := xcli.NewCommand("list", nop)
		return xcli.NewCommand("todo", nil).Add(addCmd, listCmd)
	}

	for _, args := range [][]string{{}, {"lsit"}, {"add"}} {
		root := newRoot()
		if err := root.Parse(args...); err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(root.Run(context.Background(), ring.New()))
	}
	// Output:
	// subcommand required for `todo`
	// unknown command: `lsit`
	// command `add` requires minimum 1 argument 0 given
}

func ExampleCommand_PreRun() {
	setup := func(
		ctx context.Context,
		ccx *xcli.CmdCtx,
	) (context.Context, error) {

		ccx.Cmd.MetaSet("file", ccx.Cmd.GetString("file"))
		return ctx, nil
	}
	add := func(_ context.Context, ccx *xcli.CmdCtx) error {
		file := ccx.Cmd.MetaGet("file").(string)
		task := strings.Join(ccx.Cmd.Args(), " ")
		_, err := fmt.Fprintf(ccx.Rng.Stdout(), "added %q to %s\n", task, file)
		return err
	}

	addCmd := xcli.NewCommand("add", add).ArgsRule(xcli.MinArgsRule(1))
	root := xcli.NewCommand("todo", nil).
		PreRun(setup).
		Add(addCmd)
	root.StringSL("file", "f", "todo.txt", "task list file")

	if err := root.Parse("-f", "work.txt", "add", "buy", "milk"); err != nil {
		fmt.Println(err)
		return
	}
	if err := root.Run(context.Background(), ring.New()); err != nil {
		fmt.Println(err)
	}
	// Output:
	// added "buy milk" to work.txt
}

func ExampleDetectTTY() {
	rng := xcli.DetectTTY(ring.New())

	// Prints false here: go test pipes the example's standard output.
	fmt.Println(xcli.IsStdOutTTY(rng))
	// Output:
	// false
}
