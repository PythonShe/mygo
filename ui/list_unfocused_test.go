package ui

import (
	"fmt"
	"testing"
)

// TestListSelectionGraysWithoutFocus checks that a list shows its chosen
// row in the accent only while it has the focus, and in gray with the
// text's own color otherwise, as AppKit's tables and Sidebar: of two lists,
// the one the keys move stands out.
func TestListSelectionGraysWithoutFocus(t *testing.T) {
	a, b := 1, 2
	sa, sb := ListState{Selected: &a}, ListState{Selected: &b}
	rows := map[string]*node{}
	var focusedA, selectedA bool
	var theme *Theme
	tt := coreNewTester(func(c *context) {
		theme = c.theme
		coreRow(c).Fill().AlignItems(Stretch).Children(func() {
			coreList(c, &sa, 5, func(i int) {
				rows[fmt.Sprintf("a%d", i)] = coreTextf(c, "a%d", i).parent
				if i == 1 {
					focusedA, selectedA = c.row.f.owner.FocusWithin(), c.row.chosen
				}
			}).Width(150)
			coreList(c, &sb, 5, func(i int) {
				rows[fmt.Sprintf("b%d", i)] = coreTextf(c, "b%d", i).parent
			}).Width(150)
		})
	}, 300, 300)
	accent := func(name string) bool {
		r := rows[name]
		return r.checked == 2 && r.bg == theme.Accent && r.ts.set&setColor != 0 && r.ts.color == theme.AccentText
	}
	gray := func(name string) bool {
		r := rows[name]
		return r.checked == 2 && r.bg == theme.SurfacePressed && r.ts.set&setColor == 0
	}
	if !gray("a1") || !gray("b2") {
		t.Fatal("chosen rows of lists without the focus show in the accent")
	}
	if err := tt.Click("a1"); err != nil {
		t.Fatal(err)
	}
	if !accent("a1") || !gray("b2") {
		t.Fatalf("list a has the focus: a1 accent %v, b2 gray %v", accent("a1"), gray("b2"))
	}
	if !focusedA || !selectedA {
		t.Errorf("a row of the focused list reads focused %v, chosen %v", focusedA, selectedA)
	}
	// A window drawn inactive fades the focused list's accent, as before.
	tt.SetWindowActive(false)
	if r := rows["a1"]; r.bg != theme.inactiveAccent() || r.ts.color != theme.AccentText {
		t.Errorf("inactive window: a1 shows %v under %v", r.bg, r.ts.color)
	}
	tt.SetWindowActive(true)
	if err := tt.Click("b2"); err != nil {
		t.Fatal(err)
	}
	if !gray("a1") || !accent("b2") {
		t.Fatalf("list b has the focus: a1 gray %v, b2 accent %v", gray("a1"), accent("b2"))
	}
	if focusedA {
		t.Error("a row of the list without the focus reads it focused")
	}
	// The gray row takes no hover color over it, as painted.
	r := rows["a1"].st
	tt.Move(r.x+r.w-6, r.y+r.h/2)
	if c := pixel(tt.Image(), r.x+r.w-6, r.y+r.h/2); c != theme.SurfacePressed {
		t.Errorf("hovered: a1 paints %v, want %v", c, theme.SurfacePressed)
	}
}

// TestListAccentFollowsFocusWithin checks that a chosen row keeps the
// accent while the focus is on something inside the list, as a field in a
// row, or in a menu opened from one of its rows.
func TestListAccentFollowsFocusWithin(t *testing.T) {
	sel, other, dialog, name := 1, "", false, ""
	text := []string{"r0", "r1", "r2"}
	s := ListState{Selected: &sel}
	var row1 *node
	var theme *Theme
	tt := coreNewTester(func(c *context) {
		theme = c.theme
		coreColumn(c).Fill().Children(func() {
			coreTextInput(c, &other).Label("other")
			coreList(c, &s, 3, func(i int) {
				coreRow(c).Children(func() {
					coreTextf(c, "row %d", i)
					coreTextInput(c, &text[i]).Label(fmt.Sprintf("field %d", i)).Width(80)
					corePopoverMenuButton(c, fmt.Sprintf("menu %d", i), func(m *Menu) { m.Item("Rename") })
				})
				if i == 1 {
					row1 = c.parent
				}
			}).Grow(1)
		})
		coreModal(c, &dialog, func() {
			coreTextInput(c, &name).Label("name").AutoFocus()
		})
	}, 400, 300)
	tt.SetPreferences(Preferences{ReduceMotion: true, TextScale: 1})
	if err := tt.Click("other"); err != nil {
		t.Fatal(err)
	}
	if row1.bg != theme.SurfacePressed {
		t.Fatalf("focus outside: r1 shows %v", row1.bg)
	}
	if err := tt.Click("field 1"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("field 1") || row1.bg != theme.Accent {
		t.Fatalf("focus in a row's field: r1 shows %v (field focused %v)", row1.bg, tt.Focused("field 1"))
	}
	if err := tt.Click("menu 1"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Rename") {
		t.Fatalf("no menu; texts %q", tt.Texts())
	}
	tt.Key(0, KeyDown) // the menu takes the focus
	tt.Frame()
	if row1.bg != theme.Accent {
		t.Errorf("a menu opened from a row: r1 shows %v", row1.bg)
	}
	// A dialog opened from the list takes the keys, as AppKit's sheets do:
	// the list goes gray.
	tt.Key(0, KeyEscape)
	if err := tt.Click("row 1"); err != nil {
		t.Fatal(err)
	}
	dialog = true
	tt.Frame()
	tt.Frame()
	if !tt.Focused("name") || row1.bg != theme.SurfacePressed {
		t.Errorf("a dialog opened from the list: name focused %v, r1 shows %v", tt.Focused("name"), row1.bg)
	}
}

// TestTableRowsAndOutlineGrayWithoutFocus checks the chosen row of a
// Table that chooses rows, and the arrow of an Outline's chosen row, which
// stays muted on the gray rather than in the accent's text.
func TestTableRowsAndOutlineGrayWithoutFocus(t *testing.T) {
	tsel, osel, value := 1, 0, ""
	ts := &ListState{Selected: &tsel}
	var os OutlineState[string]
	os.List.Selected = &osel
	var trow, orow *node
	var theme *Theme
	tt := coreNewTester(func(c *context) {
		theme = c.theme
		coreColumn(c).Fill().Children(func() {
			coreTextInput(c, &value).Label("field")
			coreTable(c, ts, []TableColumn{{Title: "Name"}}, 3, func(row, col int) {
				n := coreTextf(c, "t%d", row)
				if row == 1 {
					trow = c.row.e
				}
				_ = n
			}).Height(120)
			coreOutline(c, &os, []string{"src", "docs"}, children, func(path string) {
				coreText(c, baseName(path))
				if path == "src" {
					orow = c.row.e
				}
			}).Grow(1)
		})
	}, 400, 400)
	if err := tt.Click("t1"); err != nil {
		t.Fatal(err)
	}
	if trow.bg != theme.Accent {
		t.Fatalf("focused table: its row shows %v", trow.bg)
	}
	if err := tt.Click("src"); err != nil {
		t.Fatal(err)
	}
	if trow.bg != theme.SurfacePressed || orow.bg != theme.Accent || arrowColor(theme, orow, os.List.frame.emphasized()) != theme.AccentText {
		t.Fatalf("focused outline: table row %v, outline row %v, arrow %v", trow.bg, orow.bg, arrowColor(theme, orow, os.List.frame.emphasized()))
	}
	if err := tt.Click("field"); err != nil {
		t.Fatal(err)
	}
	if orow.bg != theme.SurfacePressed || arrowColor(theme, orow, os.List.frame.emphasized()) != theme.TextMuted {
		t.Errorf("outline without the focus: row %v, arrow %v", orow.bg, arrowColor(theme, orow, os.List.frame.emphasized()))
	}
}

// TestTableCellsGrayWithoutFocus checks that the block of a table
// choosing cells shows in gray, with the text's own color, while the
// table does not have the focus, and that CellsStyle's colors stay.
func TestTableCellsGrayWithoutFocus(t *testing.T) {
	s := &ListState{CellSelect: true}
	sel := &Selection[int]{}
	s.Selection = sel
	cols := []TableColumn{{Title: "#", RowHeader: true, Width: 40, Fixed: true}, {Title: "A"}, {Title: "B"}}
	value := ""
	var cell *node
	var theme *Theme
	tt := coreNewTester(func(c *context) {
		theme = c.theme
		coreColumn(c).Fill().Children(func() {
			coreTextInput(c, &value).Label("field")
			coreTable(c, s, cols, 4, func(row, col int) {
				n := coreTextf(c, "r%dc%d", row, col)
				if row == 1 && col == 2 {
					cell = n.parent
				}
			}).Grow(1)
		})
	}, 400, 300)
	if err := tt.Click("r1c2"); err != nil {
		t.Fatal(err)
	}
	if cell.bg != theme.Accent || cell.ts.color != theme.AccentText {
		t.Fatalf("focused table: the cell shows %v under %v", cell.bg, cell.ts.color)
	}
	if err := tt.Click("field"); err != nil {
		t.Fatal(err)
	}
	if cell.bg != theme.SurfacePressed || cell.ts.set&setColor != 0 || cell.checked != 2 {
		t.Fatalf("table without the focus: the cell shows %v, its text set %v, checked %d", cell.bg, cell.ts.set&setColor != 0, cell.checked)
	}
	s.CellsStyle = CellsStyle{Cell: Hex("#ffeeaa"), CellText: Hex("#112233")}
	tt.Frame()
	if cell.bg != s.CellsStyle.Cell || cell.ts.color != s.CellsStyle.CellText {
		t.Errorf("CellsStyle without the focus: the cell shows %v under %v", cell.bg, cell.ts.color)
	}
}
