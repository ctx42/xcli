[![Go](https://github.com/ctx42/xcli/actions/workflows/go.yml/badge.svg)](https://github.com/ctx42/xcli/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ctx42/xcli.svg)](https://pkg.go.dev/github.com/ctx42/xcli)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ctx42/xcli)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

# xcli

A small framework for Go command line programs with nested subcommands.

xcli builds a program as a tree of commands. Each command has its own flag set,
help text, argument rule, and function. You parse the arguments once, and xcli
finds the command to run and runs it with a context. Flags come from
[xflag](https://github.com/ctx42/xflag), so required flags and long/short
aliases work out of the box.

Commands never touch `os.Stdout`, `os.Stderr`, or the process environment
directly. They get them through a [ring](https://github.com/ctx42/ring)
passed in at run time. Tests can hand a command captured streams and a fake
environment, and the `xclitest` package wraps that setup for you.

## Features

- Nested subcommands, each with its own flags, description, and examples.
- Generated help for every command, shown on `-h` or `--help` even when the
  command does not define a help flag.
- Argument count rules (`ArgsRule`, `MinArgsRule`) checked before a command
  runs.
- Pre-run hooks run from the root command down to the selected one.
- Metadata shared by every command in the tree.
- Clear errors for a missing or mistyped subcommand.
- `Execute` and `RunCmd` turn panics into errors and cancel the context on
  `SIGINT` and `SIGTERM`.
- Terminal detection for standard output and standard error.
- Test helpers in `xclitest` with captured streams and ready-made command
  functions.

## Install

```shell
go get github.com/ctx42/xcli
```

Import the packages you need:

```go
import (
	"github.com/ctx42/xcli/pkg/xcli"
	"github.com/ctx42/xcli/pkg/xcli/xclitest" // In tests only.
)
```

## Usage

### Quickstart

A single command with a help flag and a `--name` flag:

<!-- gmmce:pkg/xcli/ExampleCommand -->
```go
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
```

`UsageFlagName` tells the command which flag requests help. On `-h` or
`--help`, `Execute` prints the help to the standard error of the ring
instead of running the command function. The same text comes from `Help`:

<!-- gmmce:pkg/xcli/ExampleCommand_Help -->
```go
cmd := xcli.NewCommand("hello", nil).
	ShortDesc("print a greeting").
	UsageFlagName("help")
cmd.Flags().BoolSL("help", "h", false, "show help")
cmd.Flags().StringSL("name", "n", "World", "who to greet")

// Execute prints the same text to standard error on -h or --help.
fmt.Print(cmd.Help())
// Output:
// Usage of `hello`:
//
//   print a greeting
//
// OPTIONS:
//   -h, --help    show help
//   -n, --name    who to greet
```

### Subcommands

Build the tree bottom-up with `Add`. A root command with no function of its
own only dispatches to its subcommands. Flags belong to the command they are
defined on and go before its subcommand: `todo -f work.txt list`.

<!-- gmmce:pkg/xcli/ExampleCommand_Add -->
```go
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
//   -f, --file    task list file
//   -h, --help    show help
//
// COMMANDS:
//   add           add a task
//   list          list tasks
```

A missing subcommand, a mistyped one, and a broken argument rule each return
an error that names the problem:

<!-- gmmce:pkg/xcli/ExampleCommand_Execute_errors -->
```go
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
```

The errors wrap the `ErrSubCmdReq` and `ErrUnkCommand` sentinels, so callers
can match them with `errors.Is`. Without `ArgsRule`, a command accepts no
arguments.

### Pre-run hooks and metadata

Pre-run hooks run before the command function, starting at the root and
ending at the selected command. A hook may return a new context, which the
rest of the chain and the command receive. Metadata set on any command is
shared by the whole tree, which makes it a simple way to pass values from a
hook on the root to a subcommand:

<!-- gmmce:pkg/xcli/ExampleCommand_PreRun -->
```go
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
```

A subcommand can also read a flag of any ancestor directly with
`ccx.Cmd.GetFlagSet("todo").GetString("file")`.

### Terminal detection

`DetectTTY` records in the ring metadata whether standard output and standard
error are terminals. Read the result with `IsStdOutTTY` and `IsStdErrTTY`:

<!-- gmmce:pkg/xcli/ExampleDetectTTY -->
```go
rng := xcli.DetectTTY(ring.New())

// Prints false here: go test pipes the example's standard output.
fmt.Println(xcli.IsStdOutTTY(rng))
// Output:
// false
```

### Testing commands

`xclitest.Tester` builds a ring with captured streams and fails the test when
output appears on a stream nobody expected. `Ctx` parses the arguments and
returns the command context to call the function with. Here `greet` is the
quickstart function declared at package level:

```go
func Test_greet(t *testing.T) {
	tst := xclitest.New(t).WetStdout()
	cmd := xcli.NewCommand("hello", greet)
	cmd.Flags().StringSL("name", "n", "World", "who to greet")
	ccx := tst.Ctx(cmd, "--name", "gopher")

	err := greet(t.Context(), ccx)

	assert.NoError(t, err)
	assert.Equal(t, "Hello, gopher!\n", tst.Stdout())
}
```

`assert` comes from [ctx42/testing](https://github.com/ctx42/testing).
`xclitest` also ships ready-made command functions: `DbgFunc` prints the
command name and arguments, and `NopFunc` does nothing.

## License

xcli is released under the MIT license. See [LICENSE.md](LICENSE.md).
