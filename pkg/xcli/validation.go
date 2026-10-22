// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"fmt"
)

var _ Rule = RuleFunc(nil)

// Rule represents a validation rule.
type Rule interface {
	// Validate validates a value and returns an error if validation fails.
	Validate(v any) error
}

// RuleFunc is an adapter allowing the use of ordinary functions as [Rule].
type RuleFunc func(v any) error

// Validate calls fn(v).
func (fn RuleFunc) Validate(v any) error { return fn(v) }

// ArgsRule requires command to have n arguments. Its error wraps
// [ErrArgCount].
func ArgsRule(n int) Rule {
	return RuleFunc(func(v any) error {
		vv, ok := v.([]string)
		if !ok {
			return ErrInvArgType
		}
		if len(vv) != n {
			if n == 0 {
				return fmt.Errorf("%w: want none, got %d", ErrArgCount, len(vv))
			}
			const format = "%w: want %d, got %d"
			return fmt.Errorf(format, ErrArgCount, n, len(vv))
		}
		return nil
	})
}

// MinArgsRule requires command to have minimum n arguments. Its error wraps
// [ErrArgCount].
func MinArgsRule(n int) Rule {
	return RuleFunc(func(v any) error {
		vv, ok := v.([]string)
		if !ok {
			return ErrInvArgType
		}
		if len(vv) < n {
			const format = "%w: want at least %d, got %d"
			return fmt.Errorf(format, ErrArgCount, n, len(vv))
		}
		return nil
	})
}
