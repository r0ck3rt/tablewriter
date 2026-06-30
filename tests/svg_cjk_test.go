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
