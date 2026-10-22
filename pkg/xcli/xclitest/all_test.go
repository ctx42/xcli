// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"sort"
)

// Sort sorts the passed slice and returns it.
func Sort(in []string) []string { sort.Strings(in); return in }
