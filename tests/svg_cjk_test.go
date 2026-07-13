package tests

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
)

var svgRootWidthRe = regexp.MustCompile(`<svg[^>]*\bwidth="([0-9.]+)"`)

// svgRootWidth renders a single-column SVG table whose only cell is content and
// returns the width of the root <svg> element.
func svgRootWidth(t *testing.T, content string) float64 {
	t.Helper()
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	table.Append([]string{content})
	table.Render()

	m := svgRootWidthRe.FindStringSubmatch(buf.String())
	if m == nil {
		t.Fatalf("no width attribute found in SVG output:\n%s", buf.String())
	}
	w, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("failed to parse width %q: %v", m[1], err)
	}
	return w
}

// TestSVGWideRuneColumnWidth verifies that the SVG renderer sizes columns by
// the display width of the content, so that wide (CJK) runes which occupy two
// terminal cells are not packed into half the space they need. A cell of four
// fullwidth runes has the same display width as eight ASCII characters and so
// must produce an equally wide column. Sizing by rune count would make the CJK
// column too narrow and the text would overflow its cell.
func TestSVGWideRuneColumnWidth(t *testing.T) {
	wide := svgRootWidth(t, "ああああ")      // 4 fullwidth runes, display width 8
	ascii := svgRootWidth(t, "abcdefgh") // 8 ASCII runes, display width 8

	if wide != ascii {
		t.Errorf("CJK column sized by rune count, not display width: width(ああああ)=%v, width(abcdefgh)=%v; want equal", wide, ascii)
	}
}

// TestSVGMixedWidthColumnWidth verifies that a cell mixing ASCII and fullwidth
// runes is sized by the sum of their display widths rather than the rune count.
// "aあ" is one ASCII cell plus one fullwidth (two-cell) rune, a display width of
// three, the same as three ASCII characters. Sizing by rune count would treat
// "aあ" as two runes and make the column too narrow.
func TestSVGMixedWidthColumnWidth(t *testing.T) {
	mixed := svgRootWidth(t, "aあ")  // display width 1 + 2 = 3
	ascii := svgRootWidth(t, "abc") // display width 3

	if mixed != ascii {
		t.Errorf("mixed cell sized by rune count, not display width: width(aあ)=%v, width(abc)=%v; want equal", mixed, ascii)
	}
}

// TestSVGWideRuneColumnWidthScales verifies the display-width sizing holds at
// several lengths, so a run of n fullwidth runes matches 2n ASCII characters
// and not only at one particular length.
func TestSVGWideRuneColumnWidthScales(t *testing.T) {
	cases := []struct {
		wide  string
		ascii string
	}{
		{"中", "ab"},
		{"中文", "abcd"},
		{"中文字", "abcdef"},
	}
	for _, c := range cases {
		wide := svgRootWidth(t, c.wide)
		ascii := svgRootWidth(t, c.ascii)
		if wide != ascii {
			t.Errorf("width(%q)=%v, width(%q)=%v; want equal", c.wide, wide, c.ascii, ascii)
		}
	}
}

// TestSVGCombiningMarkColumnWidth verifies that combining marks, which occupy no
// display cells of their own, do not widen a column. "é" (e followed by a
// combining acute accent) has the display width of a single "e" even though it
// is two runes; sizing by rune count would make the column too wide.
func TestSVGCombiningMarkColumnWidth(t *testing.T) {
	combining := svgRootWidth(t, "é") // e + combining acute, display width 1
	plain := svgRootWidth(t, "e")      // display width 1

	if combining != plain {
		t.Errorf("combining mark widened the column: width(e+U+0301)=%v, width(e)=%v; want equal", combining, plain)
	}
}
