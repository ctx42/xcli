// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"github.com/ctx42/ring/pkg/ring"
)

// CmdCtx represents command context.
type CmdCtx struct {
	Rng *ring.Ring // Command environment.
	Cmd *Command   // The root command.
}

// NewCmdCtx reruns a new instance of command context.
func NewCmdCtx(rng *ring.Ring, cmd *Command) *CmdCtx {
	return &CmdCtx{Rng: rng, Cmd: cmd}
}
