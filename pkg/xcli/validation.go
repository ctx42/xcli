// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"fmt"
)

// Rule represents a validation rule.
type Rule interface {
	// Validate validates a value and returns an error if validation fails.
	Validate(v any) error
}

// RuleFunc is an adapter allowing the use of ordinary functions as [Rule].
type RuleFunc func(v any) error

// Validate calls fn(v).
func (fn RuleFunc) Validate(v any) error { return fn(v) }

// ArgsRule requires command to have n arguments.
func ArgsRule(n int) Rule {
	return RuleFunc(func(v any) error {
		vv, ok := v.([]string)
		if !ok {
			return ErrInvArgType
		}
		if len(vv) != n {
			if n == 0 {
				return fmt.Errorf("requires no arguments %d given", len(vv))
			}
			p := "arguments"
			if n == 1 {
				p = "argument"
			}
			return fmt.Errorf("requires %d %s %d given", n, p, len(vv))
		}
		return nil
	})
}

// MinArgsRule requires command to have minimum n arguments.
func MinArgsRule(n int) Rule {
	return RuleFunc(func(v any) error {
		vv, ok := v.([]string)
		if !ok {
			return ErrInvArgType
		}
		if len(vv) < n {
			p := "arguments"
			if n == 1 {
				p = "argument"
			}
			return fmt.Errorf("requires minimum %d %s %d given", n, p, len(vv))
		}
		return nil
	})
}
