// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"context"
	"fmt"

	"github.com/ctx42/ring/pkg/ring"

	"github.com/ctx42/xcli/pkg/xcli"
)

func ExampleCommand() {
	greet := func(_ context.Context, ccx *xcli.CmdCtx) error {
		const format = "%s: hello %s from %s\n"
		cmd := ccx.Cmd
		name := cmd.GetFlagSet("app").GetString("name")
		out := ccx.Rng.Stdout()
		_, err := fmt.Fprintf(out, format, cmd.FullName(), cmd.Args(), name)
		return err
	}

	sub := xcli.NewCommand("greet", greet).
		ShortDesc("print a greeting").
		ArgsRule(xcli.MinArgsRule(1))
	root := xcli.NewCommand("app", nil).Add(sub)
	root.String("name", "xcli", "name of the greeter")

	if err := root.Parse("--name", "gopher", "greet", "world"); err != nil {
		fmt.Println(err)
		return
	}
	if err := xcli.RunCmd(context.Background(), ring.New(), root); err != nil {
		fmt.Println(err)
	}
	// Output:
	// app greet: hello [world] from gopher
}
