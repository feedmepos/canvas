package text

import (
	"github.com/tdewolff/font"
)

// outlineTables are the SFNT tables that hold glyph outlines and their
// variations. The shaper only reads them to compute glyph extents, which
// HarfBuzz uses for fallback mark positioning when a font has no GPOS mark
// attachment for a mark.
var outlineTables = map[string]bool{
	"glyf": true,
	"loca": true,
	"gvar": true,
	"CFF ": true,
	"CFF2": true,
}

// NewShaperSFNTWithoutOutlines returns a text shaper for sfnt that is built from
// a copy of the font without glyph outline tables. Parsing the outlines
// dominates the shaper's memory for large fonts (tens of megabytes for a CJK
// face), while shaping results only depend on them for fallback mark
// positioning. Fonts whose marks are positioned through GPOS shape
// identically. If the shaper cannot be built without outlines, the complete
// font is used instead.
func NewShaperSFNTWithoutOutlines(sfnt *font.SFNT) (Shaper, error) {
	lean := &font.SFNT{
		IsCFF:      sfnt.IsCFF,
		IsTrueType: sfnt.IsTrueType,
		Tables:     make(map[string][]byte, len(sfnt.Tables)),
	}
	for tag, table := range sfnt.Tables {
		if !outlineTables[tag] {
			lean.Tables[tag] = table
		}
	}
	if len(lean.Tables) == len(sfnt.Tables) {
		return NewShaperSFNT(sfnt)
	}
	if shaper, err := NewShaper(lean.Write(), 0); err == nil {
		return shaper, nil
	}
	return NewShaperSFNT(sfnt)
}
