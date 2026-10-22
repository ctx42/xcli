// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ctx42/ring/pkg/ring"
	"github.com/mattn/go-isatty"
)

// defaultFunc is default [Func] if none provided.
func defaultFunc(_ context.Context, ccx *CmdCtx) error {
	return fmt.Errorf("%w for %#q", ErrNoDefault, ccx.Cmd.Name())
}

// Marshal marshals v to string.
//
// If v implements Marshaler or MarshalText interfaces, it will be used.
func Marshal(v any) ([]byte, error) {
	if m, ok := v.(Marshaler); ok {
		return m.MarshalCLI()
	}
	txt := fmt.Sprintf("%s", v)
	return []byte(txt), nil
}

// RunCmd executes command and recovers from panics. If the error returns an
// error, we always log it to the standard error and to the log configured in
// [ring.Ring].
func RunCmd(ctx context.Context, rng *ring.Ring, cmd *Command) (err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		switch e := rec.(type) {
		case string:
			err = errors.New(e)
		case error:
			err = e
		case fmt.Stringer:
			err = errors.New(e.String())
		default:
			err = fmt.Errorf("do not know how to handle `%T` panics", rec)
		}
	}()

	ctx, cxl := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer cxl()
	return cmd.Run(ctx, rng)
}

// DetectTTY detects if standard output and error are terminals and sets
// [MetaKeySOTerm] and [MetaKeySETerm] accordingly. If the [MetaKeySOTerm] and
// [MetaKeySETerm] are already set, it will not change their values.
func DetectTTY(rng *ring.Ring) *ring.Ring {
	if val := rng.MetaGet(MetaKeySOTerm); val == nil {
		var soTTY bool
		if sout, ok := rng.Stdout().(*os.File); ok {
			soTTY = isatty.IsTerminal(sout.Fd())
		}
		rng.MetaSet(MetaKeySOTerm, soTTY)
	}

	if val := rng.MetaGet(MetaKeySETerm); val == nil {
		var seTTY bool
		if eout, ok := rng.Stderr().(*os.File); ok {
			seTTY = isatty.IsTerminal(eout.Fd())
		}
		rng.MetaSet(MetaKeySETerm, seTTY)
	}
	return rng
}

// IsStdOutTTY returns true if stdout defined in the ring is a terminal.
func IsStdOutTTY(rng *ring.Ring) bool {
	if val := rng.MetaGet(MetaKeySOTerm); val != nil {
		if is, ok := val.(bool); ok {
			return is
		}
	}
	return false
}

// IsStdErrTTY returns true if stderr defined in the ring is a terminal.
func IsStdErrTTY(rng *ring.Ring) bool {
	if val := rng.MetaGet(MetaKeySETerm); val != nil {
		if is, ok := val.(bool); ok {
			return is
		}
	}
	return false
}
