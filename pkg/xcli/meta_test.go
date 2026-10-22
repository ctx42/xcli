// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli_test

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"

	. "github.com/ctx42/xcli/pkg/xcli"
)

func Test_Meta_MetaLookup(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		met := NewMeta().MetaSet("key", 1)

		// --- When ---
		have, ok := met.MetaLookup("key")

		// --- Then ---
		assert.True(t, ok)
		assert.Equal(t, 1, have)
	})

	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		met := NewMeta()

		// --- When ---
		have, ok := met.MetaLookup("key")

		// --- Then ---
		assert.False(t, ok)
		assert.Nil(t, have)
	})
}

func Test_Meta_MetaGet(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		met := NewMeta().MetaSet("key", 1)

		// --- When ---
		have := met.MetaGet("key")

		// --- Then ---
		assert.Equal(t, 1, have)
	})

	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		met := NewMeta()

		// --- When ---
		have := met.MetaGet("key")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_Meta_MetaDelete(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		met := NewMeta().MetaSet("key", 1).MetaSet("other", 2)

		// --- When ---
		met.MetaDelete("key")

		// --- Then ---
		assert.Equal(t, map[string]any{"other": 2}, met.MetaAll())
	})

	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		met := NewMeta().MetaSet("other", 2)

		// --- When ---
		met.MetaDelete("key")

		// --- Then ---
		assert.Equal(t, map[string]any{"other": 2}, met.MetaAll())
	})
}
