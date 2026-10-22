// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"errors"
)

// ErrTest is a general error used in tests.
var ErrTest = errors.New("test error")

// Testing exports.
var (
	DefaultFunc = defaultFunc
)

// Describe is exporting [Command.describe] for testing.
func (cmd *Command) Describe() string { return cmd.describe() }

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
