// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

// Meta represents metadata collection.
type Meta struct {
	m map[string]any
}

// NewMeta returns new [Meta] instance.
func NewMeta() Meta { return Meta{m: make(map[string]any, 10)} }

// MetaLookup returns the value of the metadata variable named by the key. If
// the variable is present in the metadata collection the value (which may be
// empty or nil) is returned and the boolean is true. Otherwise, the returned
// value will be nil and the boolean will be false.
func (m Meta) MetaLookup(key string) (any, bool) {
	val, ok := m.m[key]
	return val, ok
}

// MetaGet returns the value of the metadata variable named by the key. If the
// variable is not present in the metadata collection nil is returned.
func (m Meta) MetaGet(key string) any { return m.m[key] }

// MetaSet sets the value of metadata variable named by the key.
func (m Meta) MetaSet(key string, value any) Meta {
	m.m[key] = value
	return m
}

// MetaDelete deletes the metadata entry identified by the key.
func (m Meta) MetaDelete(key string) { delete(m.m, key) }

// MetaAll returns underlying map used for metadata collection.
func (m Meta) MetaAll() map[string]any { return m.m }
