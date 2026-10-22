// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package xcli is a framework for building command line programs.
//
// A program is a tree of [Command] values. Each command has a name, a flag
// set, an optional [Func] holding its logic, and subcommands added with
// [Command.Add]. [Command.Execute] parses the arguments, finds the command to
// run, and runs it after the [PreRun] functions of the commands from the root
// down to it. Commands reach the program environment
// and streams through the [github.com/ctx42/ring/pkg/ring.Ring] in their
// [CmdCtx].
//
//	sub := xcli.NewCommand("version", versionFunc).
//		ShortDesc("print the version")
//	root := xcli.NewCommand("app", appFunc).
//		UsageFlagName("help").
//		Add(sub)
//	root.Flags().BoolSL("help", "h", false, "show help")
//
//	return root.Execute(ctx, ring.New())
package xcli

import (
	"context"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Func is a signature for CLI command logic.
type Func func(ctx context.Context, ccx *CmdCtx) error

// Metadata [ring.Ring] keys.
const (
	// MetaKeySOTerm is the metadata key of a boolean value which is true when
	// the program's standard output is a terminal.
	MetaKeySOTerm = "cli-meta-sout-term"

	// MetaKeySETerm is the metadata key of a boolean value which is true when
	// the program's standard error is a terminal.
	MetaKeySETerm = "cli-meta-eout-term"

	// MetaKeyColor is the metadata key of a boolean value which is true when
	// the program can use terminal colors. [DetectTTY] sets it.
	MetaKeyColor = "cli-meta-color"
)

// PanicStackKey is the error metadata key under which [RunCmd] stores the
// stack trace of a recovered panic.
const PanicStackKey = "panic-stack"

// CLI related errors.
var (
	// ErrNoDefault is returned when a command has no function to run.
	ErrNoDefault = xrr.New("default command not defined", "ECCmdNoDefault")

	// ErrSubCmdReq is returned when the executed command requires a
	// subcommand.
	ErrSubCmdReq = xrr.New("subcommand required", "ECSubCmdReq")

	// ErrUnkCommand is returned when the requested command is unknown.
	ErrUnkCommand = xrr.New("unknown command", "ECUnkCommand")

	// ErrReqMeta is for command code to return when required command metadata
	// is missing; the xcli package itself never returns it.
	ErrReqMeta = xrr.New("required command metadata key", "ECCmdReqMeta")

	// ErrInvMeta is for command code to return when command metadata is
	// invalid because of its type, format or value; the xcli package itself
	// never returns it.
	ErrInvMeta = xrr.New("invalid command metadata key", "ECCmdInvMeta")

	// ErrArgCount is returned when a command gets a number of arguments its
	// [Rule] does not allow.
	ErrArgCount = xrr.New("wrong number of arguments", "ECArgCount")

	// ErrNotParsed is returned when a command runs before its arguments were
	// parsed with [Command.Parse].
	ErrNotParsed = xrr.New("command arguments not parsed", "ECNotParsed")

	// ErrInvArgType is returned when the command arguments passed to a rule
	// are not a slice of strings.
	ErrInvArgType = xrr.New(
		"invalid command argument rule usage",
		"ECInvArgType",
	)
)

// Marshaler is an interface that provides the ability to marshal an object
// to a byte slice which is suitable for display on the CLI.
type Marshaler interface {
	MarshalCLI() ([]byte, error)
}
