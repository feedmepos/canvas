//go:build !harfbuzz || js

package text

import (
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/tdewolff/font"
)

var shaperSamples = []struct {
	text   string
	script Script
	lang   string
}{
	{"Nasi Lemak Ayam x2   RM 12.50", Latin, "en"},
	{"Phở bò tái chín – Bánh mì ổ", Latin, "vi"},
	{"a\u0301 e\u0323\u0302 n\u0303 x\u0336", Latin, ""},
	{"fi ffl office, VAWA To", Latin, "en"},
}

func loadTestSFNT(t *testing.T, filename string) *font.SFNT {
	t.Helper()
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	sfnt, err := font.ParseFont(b, 0)
	if err != nil {
		t.Fatal(err)
	}
	return sfnt
}

func TestShaperWithoutOutlinesShapesIdentically(t *testing.T) {
	for _, filename := range []string{"../resources/DejaVuSerif.ttf", "../resources/unifont-13.0.05.ttf", "../resources/EBGaramond12-Regular.otf"} {
		sfnt := loadTestSFNT(t, filename)
		full, err := NewShaperSFNT(sfnt)
		if err != nil {
			t.Fatal(err)
		}
		lean, err := NewShaperSFNTWithoutOutlines(sfnt)
		if err != nil {
			t.Fatal(err)
		}
		for _, ppem := range []uint16{12, 24, 33} {
			for _, sample := range shaperSamples {
				want := full.Shape(sample.text, ppem, LeftToRight, sample.script, sample.lang, "", "")
				got := lean.Shape(sample.text, ppem, LeftToRight, sample.script, sample.lang, "", "")
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s ppem=%d %q: shaping without outlines differs\n got %+v\nwant %+v", filename, ppem, sample.text, got, want)
				}
			}
		}
	}
}

func TestShaperReusesBufferAcrossCalls(t *testing.T) {
	sfnt := loadTestSFNT(t, "../resources/DejaVuSerif.ttf")
	shaper, err := NewShaperSFNT(sfnt)
	if err != nil {
		t.Fatal(err)
	}
	fresh := Shaper{font: shaper.font}
	// Shape samples of decreasing length so that a reused buffer that is not
	// fully cleared would leak glyphs from an earlier call.
	samples := []string{"The quick brown fox jumps over the lazy dog", "office", "Vi", ""}
	for round := 0; round < 3; round++ {
		for _, s := range samples {
			want := fresh.Shape(s, 16, LeftToRight, Latin, "en", "", "")
			got := shaper.Shape(s, 16, LeftToRight, Latin, "en", "", "")
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("round %d %q: reused buffer gives %+v, fresh buffer gives %+v", round, s, got, want)
			}
		}
	}
}

func TestShaperIsSafeForConcurrentUse(t *testing.T) {
	sfnt := loadTestSFNT(t, "../resources/DejaVuSerif.ttf")
	shaper, err := NewShaperSFNT(sfnt)
	if err != nil {
		t.Fatal(err)
	}
	want := make([][]Glyph, len(shaperSamples))
	for i, sample := range shaperSamples {
		want[i] = shaper.Shape(sample.text, 24, LeftToRight, sample.script, sample.lang, "", "")
	}

	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for n := 0; n < 50; n++ {
				i := (worker + n) % len(shaperSamples)
				sample := shaperSamples[i]
				got := shaper.Shape(sample.text, 24, LeftToRight, sample.script, sample.lang, "", "")
				if !reflect.DeepEqual(got, want[i]) {
					t.Errorf("worker %d %q: concurrent shaping differs", worker, sample.text)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
}
