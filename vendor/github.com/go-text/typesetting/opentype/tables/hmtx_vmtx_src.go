// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

// https://learn.microsoft.com/en-us/typography/opentype/spec/hmtx
type Hmtx struct {
	Metrics []LongHorMetric `arrayCount:""`
	// avances are padded with the last value
	// and side bearings are given
	LeftSideBearings []int16 `arrayCount:""`
}

func (table Hmtx) IsEmpty() bool {
	return len(table.Metrics)+len(table.LeftSideBearings) == 0
}

func (table Hmtx) Advance(gid GlyphID) int16 {
	LM, LS := len(table.Metrics), len(table.LeftSideBearings)
	index := int(gid)
	if index < LM {
		return table.Metrics[index].AdvanceWidth
	} else if index < LS+LM { // return the last value
		return table.Metrics[len(table.Metrics)-1].AdvanceWidth
	}
	return 0
}

type LongHorMetric struct {
	AdvanceWidth, LeftSideBearing int16
}

// https://learn.microsoft.com/en-us/typography/opentype/spec/vmtx
type Vmtx = Hmtx
