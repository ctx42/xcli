// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"context"
	"errors"

	"github.com/ctx42/ring/pkg/ring"
)

// ErrTest is a general error used in tests.
var ErrTest = errors.New("test error")

// Testing exports.
var (
	DefaultFunc = defaultFunc
)

// Describe exports the describe method of [Command] for testing.
func (cmd *Command) Describe() string { return cmd.describe() }

// Run exports the run method of [Command] for testing.
func (cmd *Command) Run(ctx context.Context, rng *ring.Ring) error {
	return cmd.run(ctx, rng)
}

// RunPreRun exports the runPreRun function for testing.
var RunPreRun = runPreRun

// TMarshaler is a simple type implementing Marshaler interface.
type TMarshaler string

func (tm TMarshaler) MarshalCLI() ([]byte, error) { return []byte(tm), nil }

// TErrMarshaler is a simple type implementing Marshaler interface which
// returns error when calling MarshalCLI.
type TErrMarshaler string

func (tm TErrMarshaler) MarshalCLI() ([]byte, error) {
	return nil, errors.New(string(tm))
}

func (tm TErrMarshaler) Error() string { return string(tm) }
