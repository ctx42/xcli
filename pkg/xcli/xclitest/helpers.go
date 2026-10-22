// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"context"
	"fmt"

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
