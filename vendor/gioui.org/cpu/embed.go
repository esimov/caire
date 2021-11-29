// SPDX-License-Identifier: Unlicense OR MIT

package cpu

import _ "embed"

//go:embed abi.h
var ABIH []byte

//go:embed runtime.h
var RuntimeH []byte
