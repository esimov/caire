// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package iconvg implements a compact, binary format for simple vector graphics:
icons, logos, glyphs and emoji.

WARNING: THIS FORMAT IS EXPERIMENTAL AND SUBJECT TO INCOMPATIBLE CHANGES.

A longer overview is at
https://github.com/google/iconvg

The file format is specified at
https://github.com/google/iconvg/blob/main/spec/iconvg-spec.md

This package's encoder emits byte-identical output for the same input,
independent of the platform (and specifically its floating-point hardware).
*/
package iconvg

// TODO: shapes (circles, rects) and strokes? Or can we assume that authoring
// tools will convert shapes and strokes to paths?

// TODO: mark somehow that a graphic (such as a back arrow) should be flipped
// horizontally or its paths otherwise varied when presented in a Right-To-Left
// context, such as among Arabic and Hebrew text? Or should that be the
// responsibility of higher layers, selecting different IconVG graphics based
// on context, the way they would select different PNG graphics.

// TODO: hinting?
