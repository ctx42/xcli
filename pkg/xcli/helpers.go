// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"context"
	"encoding"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"syscall"

	"github.com/ctx42/ring/pkg/ring"
	"github.com/ctx42/xrr/pkg/xrr"
	"github.com/mattn/go-isatty"
)

// defaultFunc is the [Func] used when none is provided.
func defaultFunc(_ context.Context, ccx *CmdCtx) error {
	return fmt.Errorf("%w for %#q", ErrNoDefault, ccx.Cmd.Name())
}

// Marshal marshals v to a byte slice suitable for display on the CLI.
//
// It uses the [Marshaler] or [encoding.TextMarshaler] implementation of v, in
// that order. Otherwise, a byte slice, including one of a named type, is
// returned as text and any other value is formatted with the %v verb.
func Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case Marshaler:
		return val.MarshalCLI()
	case encoding.TextMarshaler:
		return val.MarshalText()
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
		return fmt.Appendf(nil, "%s", v), nil
	}
	return fmt.Appendf(nil, "%v", v), nil
}

// RunCmd runs the command selected by a successful [Command.Parse] on the
// root command and recovers from panics. A recovered panic is returned as an
// error carrying the panic stack trace in its metadata under [PanicStackKey];
// read it with [xrr.GetStr]. A panic with an error value wraps that error.
// The context passed to the command is canceled on SIGINT (Ctrl-C) and
// SIGTERM.
func RunCmd(ctx context.Context, rng *ring.Ring, cmd *Command) (err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		meta := xrr.Meta().Str(PanicStackKey, string(debug.Stack())).Option()
		switch e := rec.(type) {
		case string:
			err = xrr.New(e, meta)
		case error:
			err = xrr.New("", xrr.WithCause(e), meta)
		case fmt.Stringer:
			err = xrr.New(e.String(), meta)
		default:
			const format = "do not know how to handle `%T` panics"
			err = xrr.Newf(format, rec, meta)
		}
	}()

	ctx, cxl := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cxl()
	return cmd.run(ctx, rng)
}

// DetectTTY detects if standard output and error are terminals and sets
// [MetaKeySOTerm] and [MetaKeySETerm] accordingly. It sets [MetaKeyColor] to
// false when the NO_COLOR environment variable is set to a non-empty value,
// and to true otherwise. It does not change a key that is already set.
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

	if val := rng.MetaGet(MetaKeyColor); val == nil {
		rng.MetaSet(MetaKeyColor, rng.EnvGet("NO_COLOR") == "")
	}
	return rng
}

// IsStdOutTTY returns true if stdout defined in the ring is a terminal.
func IsStdOutTTY(rng *ring.Ring) bool {
	is, _ := rng.MetaGet(MetaKeySOTerm).(bool)
	return is
}

// IsStdErrTTY returns true if stderr defined in the ring is a terminal.
func IsStdErrTTY(rng *ring.Ring) bool {
	is, _ := rng.MetaGet(MetaKeySETerm).(bool)
	return is
}
