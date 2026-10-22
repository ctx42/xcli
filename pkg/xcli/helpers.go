// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"context"
	"encoding"
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

// Marshal marshals v to a byte slice suitable for display on the CLI.
//
// It uses the [Marshaler] or [encoding.TextMarshaler] implementation of v, in
// that order. Otherwise, a byte slice is returned as text and any other value
// is formatted with the %v verb.
func Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case Marshaler:
		return val.MarshalCLI()
	case encoding.TextMarshaler:
		return val.MarshalText()
	case []byte:
		return fmt.Appendf(nil, "%s", val), nil
	}
	return fmt.Appendf(nil, "%v", v), nil
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
