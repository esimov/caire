// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

// OS/2 and Windows Metrics Table
// See https://learn.microsoft.com/en-us/typography/opentype/spec/os2
type Os2 struct {
	Version             uint16
	XAvgCharWidth       uint16
	USWeightClass       uint16
	USWidthClass        uint16
	fSType              uint16
	YSubscriptXSize     int16
	YSubscriptYSize     int16
	YSubscriptXOffset   int16
	YSubscriptYOffset   int16
	YSuperscriptXSize   int16
	YSuperscriptYSize   int16
	YSuperscriptXOffset int16
	ySuperscriptYOffset int16
	YStrikeoutSize      int16
	YStrikeoutPosition  int16
	sFamilyClass        int16
	panose              [10]byte
	ulCharRange         [4]uint32
	achVendID           Tag
	FsSelection         uint16
	USFirstCharIndex    uint16
	USLastCharIndex     uint16
	STypoAscender       int16
	STypoDescender      int16
	STypoLineGap        int16
	usWinAscent         uint16
	usWinDescent        uint16
	HigherVersionData   []byte `arrayCount:"ToEnd"`
}

func (os *Os2) FontPage() FontPage {
	if os.Version == 0 {
		return FontPage(os.FsSelection & 0xFF00)
	}
	return FPNone
}

// See https://docs.microsoft.com/en-us/typography/legacy/legacy_arabic_fonts
// https://github.com/Microsoft/Font-Validator/blob/520aaae/OTFontFileVal/val_OS2.cs#L644-L681
type FontPage uint16

const (
	FPNone       FontPage = 0
	FPHebrew     FontPage = 0xB100 /* Hebrew Windows 3.1 font page */
	FPSimpArabic FontPage = 0xB200 /* Simplified Arabic Windows 3.1 font page */
	FPTradArabic FontPage = 0xB300 /* Traditional Arabic Windows 3.1 font page */
	FPOemArabic  FontPage = 0xB400 /* OEM Arabic Windows 3.1 font page */
	FPSimpFarsi  FontPage = 0xBA00 /* Simplified Farsi Windows 3.1 font page */
	FPTradFarsi  FontPage = 0xBB00 /* Traditional Farsi Windows 3.1 font page */
	FPThai       FontPage = 0xDE00 /* Thai Windows 3.1 font page */
)
