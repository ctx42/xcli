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
		name := ccx.Cmd.Flags().GetString("name")
		_, err := fmt.Fprintf(ccx.Rng.Stdout(), "Hello, %s!\n", name)
		return err
	}

	cmd := xcli.NewCommand("hello", greet).
		ShortDesc("print a greeting").
		UsageFlagName("help")
	cmd.Flags().BoolSL("help", "h", false, "show help")
	cmd.Flags().StringSL("name", "n", "World", "who to greet")

	// A program uses ring.New(), which takes the arguments from os.Args.
	rng := ring.New(ring.WithArgs([]string{"--name", "gopher"}))
	if err := cmd.Execute(context.Background(), rng); err != nil {
		fmt.Println(err)
	}
	// Output:
	// Hello, gopher!
}

func ExampleCommand_Help() {
	cmd := xcli.NewCommand("hello", nil).
		ShortDesc("print a greeting").
		UsageFlagName("help")
	cmd.Flags().BoolSL("help", "h", false, "show help")
	cmd.Flags().StringSL("name", "n", "World", "who to greet")

	// Execute prints the same text to standard output on -h or --help.
	fmt.Print(cmd.Help())
	// Output:
	// Usage of `hello`:
	//
	//   print a greeting
	//
	// OPTIONS:
	//   -h, --help    show help
	//   -n, --name    who to greet (default "World")
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
	root.Flags().BoolSL("help", "h", false, "show help")
	root.Flags().StringSL("file", "f", "todo.txt", "task list file")

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
	//   -f, --file    task list file (default "todo.txt")
	//   -h, --help    show help
	//
	// COMMANDS:
	//   add           add a task
	//   list          list tasks
}

func ExampleCommand_Execute_errors() {
	nop := func(context.Context, *xcli.CmdCtx) error { return nil }
	newRoot := func() *xcli.Command {
		addCmd := xcli.NewCommand("add", nop).ArgsRule(xcli.MinArgsRule(1))
		listCmd := xcli.NewCommand("list", nop)
		return xcli.NewCommand("todo", nil).Add(addCmd, listCmd)
	}

	for _, args := range [][]string{{}, {"lsit"}, {"add"}} {
		rng := ring.New(ring.WithArgs(args))
		fmt.Println(newRoot().Execute(context.Background(), rng))
	}
	// Output:
	// subcommand required for `todo`
	// unknown command: `lsit`
	// command `add`: wrong number of arguments: want at least 1, got 0
}

func ExampleCommand_PreRun() {
	setup := func(
		ctx context.Context,
		ccx *xcli.CmdCtx,
	) (context.Context, error) {

		ccx.Cmd.MetaSet("file", ccx.Cmd.Flags().GetString("file"))
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
	root.Flags().StringSL("file", "f", "todo.txt", "task list file")

	args := []string{"-f", "work.txt", "add", "buy", "milk"}
	rng := ring.New(ring.WithArgs(args))
	if err := root.Execute(context.Background(), rng); err != nil {
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
