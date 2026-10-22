// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"os"
	"time"

	"github.com/ctx42/ring/pkg/ring"
	"github.com/ctx42/ring/pkg/ring/ringtest"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/xcli/pkg/xcli"
)

// WithEnv is the [Tester] option setting the environment to use.
func WithEnv(env []string) func(*options) {
	return func(opts *options) { opts.env = env }
}

// WithEnvKey is the [Tester] option setting the environment key to value.
func WithEnvKey(key, value string) func(*options) {
	return func(opts *options) { opts.env = append(opts.env, key+"="+value) }
}

// WithTerm is the [Tester] option setting if the program runs on a terminal.
func WithTerm(onTerm bool) func(*options) {
	return func(opts *options) { opts.term = onTerm }
}

// WithClock is the option for [Tester] setting clock function.
func WithClock(clk func() time.Time) func(*options) {
	return func(opts *options) { opts.clock = clk }
}

// options represent [Tester] options
type options struct {
	env   []string         // Program environment.
	clock func() time.Time // Function returning current time in UTC.
	term  bool             // Is the program running on terminal? Default: true.
	color bool             // May the program use terminal colors? Default: false.
}

type hidRingTester = ringtest.Tester // Hide embedded field.

// Tester represents CLI test helper.
type Tester struct {
	*hidRingTester
	t tester.T
}

// New returns new instance of Tester.
func New(t tester.T, opts ...func(*options)) *Tester {
	t.Helper()
	def := &options{env: os.Environ(), clock: time.Now, term: true}
	for _, opt := range opts {
		opt(def)
	}
	meta := map[string]any{
		xcli.MetaKeySOTerm: def.term,
		xcli.MetaKeySETerm: def.term,
		xcli.MetaKeyColor:  def.color,
	}
	rt := ringtest.New(
		t,
		ring.WithEnv(def.env),
		ring.WithMeta(meta),
		ring.WithClock(def.clock),
	)
	tst := &Tester{
		hidRingTester: rt,
		t:             t,
	}
	return tst
}

// WetStdout sets the expectation that standard output will be written to.
func (tst *Tester) WetStdout() *Tester {
	tst.t.Helper()
	tst.hidRingTester.WetStdout()
	return tst
}

// WetStderr sets expectation that standard error will be written to.
func (tst *Tester) WetStderr() *Tester {
	tst.t.Helper()
	tst.hidRingTester.WetStderr()
	return tst
}

// Cmd returns command with the given name, function and arguments.
func (tst *Tester) Cmd(name string, fn xcli.Func, args ...string) *xcli.Command {
	tst.t.Helper()
	cmd := xcli.NewCommand(name, fn)
	if err := cmd.Parse(args...); err != nil {
		tst.t.Error(err)
		return nil
	}
	return cmd
}

// Ctx returns command context with the given command and [ring.Ring] based on
// [Tester] fields.
func (tst *Tester) Ctx(cmd *xcli.Command, args ...string) *xcli.CmdCtx {
	tst.t.Helper()
	if !cmd.Parsed() {
		if err := cmd.Parse(args...); err != nil {
			tst.t.Error(err)
			return nil
		}
	}
	return xcli.NewCmdCtx(tst.Ring(), cmd)
}

// CmdAndCtx returns a command and command context based on [Tester] fields.
// It returns nils when parsing the arguments fails.
func (tst *Tester) CmdAndCtx(
	name string,
	fn xcli.Func,
	args ...string,
) (*xcli.Command, *xcli.CmdCtx) {

	tst.t.Helper()
	cmd := tst.Cmd(name, fn, args...)
	if cmd == nil {
		return nil, nil
	}
	ccx := xcli.NewCmdCtx(tst.Ring(), cmd)
	return cmd, ccx
}

// Func returns [xcli.Func] printing debugging info to the standard output.
func (tst *Tester) Func() xcli.Func { return DbgFunc }

// NopFunc returns NOOP function.
func (tst *Tester) NopFunc() xcli.Func { return NopFunc }
