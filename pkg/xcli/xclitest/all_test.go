// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xclitest

import (
	"slices"
)

// sorted sorts the passed slice in place and returns it.
func sorted(in []string) []string { slices.Sort(in); return in }
