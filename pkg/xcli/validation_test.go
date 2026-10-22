// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_ArgsRule(t *testing.T) {
	t.Run("requires zero given zero", func(t *testing.T) {
		// --- When ---
		err := ArgsRule(0).Validate([]string{})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("requires zero given nil", func(t *testing.T) {
		// --- Given ---
		var args []string

		// --- When ---
		err := ArgsRule(0).Validate(args)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - requires zero given one", func(t *testing.T) {
		// --- When ---
		err := ArgsRule(0).Validate([]string{"arg0"})

		// --- Then ---
		assert.ErrorEqual(t, "requires no arguments 1 given", err)
	})

	t.Run("requires one given one", func(t *testing.T) {
		// --- Given ---
		args := []string{"arg0"}

		// --- When ---
		err := ArgsRule(1).Validate(args)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("requires multiple given multiple", func(t *testing.T) {
		// --- Given ---
		args := []string{"arg0", "arg1"}

		// --- When ---
		err := ArgsRule(2).Validate(args)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - requires one given zero", func(t *testing.T) {
		// --- When ---
		err := ArgsRule(1).Validate([]string{})

		// --- Then ---
		assert.ErrorEqual(t, "requires 1 argument 0 given", err)
	})

	t.Run("error - requires one given two", func(t *testing.T) {
		// --- When ---
		err := ArgsRule(1).Validate([]string{"arg0", "arg1"})

		// --- Then ---
		assert.ErrorEqual(t, "requires 1 argument 2 given", err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := ArgsRule(1).Validate("arg0")

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
	})
}

func Test_MinArgsRule(t *testing.T) {
	t.Run("requires minimum zero given zero", func(t *testing.T) {
		// --- When ---
		err := MinArgsRule(0).Validate([]string{})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("requires minimum zero given nil", func(t *testing.T) {
		// --- Given ---
		var args []string

		// --- When ---
		err := MinArgsRule(0).Validate(args)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("requires minimum zero given one", func(t *testing.T) {
		// --- When ---
		err := MinArgsRule(0).Validate([]string{"arg0"})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - requires minimum 1 given zero", func(t *testing.T) {
		// --- When ---
		err := MinArgsRule(1).Validate([]string{})

		// --- Then ---
		assert.ErrorEqual(t, "requires minimum 1 argument 0 given", err)
	})

	t.Run("error - requires minimum 2 given 1", func(t *testing.T) {
		// --- Given ---
		args := []string{"arg0"}

		// --- When ---
		err := MinArgsRule(2).Validate(args)

		// --- Then ---
		assert.ErrorEqual(t, "requires minimum 2 arguments 1 given", err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := MinArgsRule(1).Validate("arg0")

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
	})
}
