package metadata

import (
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

const (
	nameFontFamily         tables.NameID = 1
	nameFontSubfamily      tables.NameID = 2
	namePreferredFamily    tables.NameID = 16 // or Typographic Family
	namePreferredSubfamily tables.NameID = 17 // or Typographic Subfamily
	nameWWSFamily          tables.NameID = 21 //
	nameWWSSubfamily       tables.NameID = 22 //
)

type os2Table struct {
	USWeightClass uint16
	USWidthClass  uint16
	FsSelection   uint16
}

// fontDescriptor provides access to family and aspect
type fontDescriptor struct {
	// these tables are required both in Family
	// and Aspect
	os2   *os2Table // optional
	names tables.Name
	head  tables.Head
}

// Describe provides access to family and aspect.
// 'buffer' may be provided to reduce allocations.
func Describe(ld *loader.Loader, buffer []byte) (family string, aspect Aspect, _ []byte) {
	var desc fontDescriptor

	// load tables, all considered optional
	buffer, _ = ld.RawTableTo(loader.MustNewTag("OS/2"), buffer)
	if os2, _, err := tables.ParseOs2(buffer); err != nil {
		desc.os2 = &os2Table{
			USWeightClass: os2.USWeightClass,
			USWidthClass:  os2.USWidthClass,
			FsSelection:   os2.FsSelection,
		}
	}

	desc.head, buffer, _ = font.LoadHeadTable(ld, buffer)

	buffer, _ = ld.RawTableTo(loader.MustNewTag("name"), buffer)
	desc.names, _, _ = tables.ParseName(buffer)

	return desc.Family(), desc.Aspect(), buffer
}

// Family returns the font family name.
func (fd *fontDescriptor) Family() string {
	var family string
	if fd.os2 != nil && fd.os2.FsSelection&256 != 0 {
		family = fd.names.Name(namePreferredFamily)
		if family == "" {
			family = fd.names.Name(nameFontFamily)
		}
	} else {
		family = fd.names.Name(nameWWSFamily)
		if family == "" {
			family = fd.names.Name(namePreferredFamily)
		}
		if family == "" {
			family = fd.names.Name(nameFontFamily)
		}
	}
	return family
}

type fontMetrics struct {
	metrics tables.Hmtx
	post    tables.Post
}

func newFontMetrics(ld *loader.Loader) (out fontMetrics) {
	raw, _ := ld.RawTable(loader.MustNewTag("post"))
	out.post, _, _ = tables.ParsePost(raw)

	raw, _ = ld.RawTable(loader.MustNewTag("maxp"))
	maxp, _, _ := tables.ParseMaxp(raw)
	_, out.metrics, _ = font.LoadHmtx(ld, int(maxp.NumGlyphs))

	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func approximatelyEqual(x, y int) bool { return abs(x-y)*33 <= max(abs(x), abs(y)) }

func (fd *fontMetrics) isMonospace() bool {
	// code adapted from fontconfig

	// try the fast shortcuts
	if fd.post.IsFixedPitch != 0 {
		return true
	}

	if fd.metrics.IsEmpty() {
		// we can't be sure, so be conservative
		return false
	}

	if len(fd.metrics.Metrics) == 1 {
		return true
	}

	// directly read the advances in the 'hmtx' table
	var firstAdvance int
	for gid, metric := range fd.metrics.Metrics {
		if gid == 0 { // ignore the 'unset' glyph, which may be different
			continue
		}
		advance := int(metric.AdvanceWidth)
		if advance == 0 { // do not count zero as a proper width
			continue
		}

		if firstAdvance == 0 {
			firstAdvance = advance
			continue
		}

		if approximatelyEqual(advance, firstAdvance) {
			continue
		}

		// two distinct advances : the font is not monospace
		return false
	}

	return true
}

// Description provides font metadata.
type Description struct {
	Family      string
	Aspect      Aspect
	IsMonospace bool
}

// Metadata queries the family and the aspect properties of the
// font loaded under [font]
func Metadata(font *loader.Loader) Description {
	var out Description

	out.Family, out.Aspect, _ = Describe(font, nil)

	metrics := newFontMetrics(font)
	out.IsMonospace = metrics.isMonospace()

	return out
}
