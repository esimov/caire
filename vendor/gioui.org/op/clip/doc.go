// SPDX-License-Identifier: Unlicense OR MIT

/*
Package clip provides operations for defining areas that applies to operations
such as paints and pointer handlers.

The current clip is initially the infinite set. Pushing an Op sets the clip
to the intersection of the current clip and pushed clip area. Popping the
area restores the clip to its state before pushing.

General clipping areas are constructed with Path. Common cases such as
rectangular clip areas also exist as convenient constructors.
*/
package clip
