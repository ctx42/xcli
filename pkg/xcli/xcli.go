// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"context"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Func is a signature for CLI command logic.
type Func func(ctx context.Context, ccx *CmdCtx) error

// Metadata [ring.Ring] keys.
const (
	// MetaKeySOTerm is metadata key which points to boolean value which is set
	// to true if program's standard output is a terminal.
	MetaKeySOTerm = "cli-meta-sout-term"

	// MetaKeySETerm is metadata key which points to boolean value which is set
	// to true if program's standard error is a terminal.
	MetaKeySETerm = "cli-meta-eout-term"

	// MetaKeyColor is a boolean metadata key indicting program can use terminal
	// colors.
	MetaKeyColor = "cli-meta-color"
)

// CLI related errors.
var (
	// ErrNoDefault is returned when no default command was defined.
	ErrNoDefault = xrr.New("default command not defined", "ECCmdNoDefault")

	// ErrSubCmdReq is returned when executed command requires subcommand.
	ErrSubCmdReq = xrr.New("subcommand required", "ECSubCmdReq")

	// ErrUnkCommand is returned when requested command is unknown.
	ErrUnkCommand = xrr.New("unknown command", "ECUnkCommand")

	// ErrReqMeta is returned when required command metadata is messing.
	ErrReqMeta = xrr.New("required command metadata key", "ECCmdReqMeta")

	// ErrInvMeta is returned when command metadata is invalid either
	// because it is of wrong type, format or value.
	ErrInvMeta = xrr.New("invalid command metadata key", "ECCmdInvMeta")

	// ErrInvArgType is error returned when command argument slice is not
	// slice of strings.
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
