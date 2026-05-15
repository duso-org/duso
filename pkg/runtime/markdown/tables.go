package markdown

import "strings"

// convertTables walks the block tree and converts any paragraph whose first
// two lines form a GFM table header + separator into a BlockTable. Any
// trailing non-table lines remain as a paragraph after the table.
func convertTables(b *Block, opts Options) {
	if !opts.Tables || b == nil {
		return
	}
	for i := 0; i < len(b.Children); i++ {
		c := b.Children[i]
		if c.Kind == BlockParagraph {
			if table, leftover := tryConvertTable(c.Text); table != nil {
				b.Children[i] = table
				if leftover != "" {
					newP := &Block{Kind: BlockParagraph, Text: leftover}
					b.Children = append(b.Children, nil)
					copy(b.Children[i+2:], b.Children[i+1:])
					b.Children[i+1] = newP
					i++ // skip inserted paragraph
				}
			}
		}
		convertTables(b.Children[i], opts)
	}
}

func tryConvertTable(text string) (*Block, string) {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return nil, ""
	}
	header := strings.TrimSpace(lines[0])
	separator := strings.TrimSpace(lines[1])
	if !strings.Contains(header, "|") || !strings.Contains(separator, "|") {
		return nil, ""
	}
	hcells := splitTableRow(header)
	aligns, ok := parseSeparator(separator)
	if !ok || len(aligns) != len(hcells) {
		return nil, ""
	}
	table := &Block{Kind: BlockTable, Aligns: aligns}
	hdr := &Block{Kind: BlockTableRow, IsHeader: true}
	for _, c := range hcells {
		hdr.Children = append(hdr.Children, &Block{Kind: BlockTableCell, Text: c})
	}
	table.Children = []*Block{hdr}

	leftover := ""
	i := 2
	for ; i < len(lines); i++ {
		row := strings.TrimSpace(lines[i])
		if row == "" || !strings.Contains(row, "|") {
			break
		}
		cells := splitTableRow(row)
		// Pad/truncate to align count.
		for len(cells) < len(aligns) {
			cells = append(cells, "")
		}
		if len(cells) > len(aligns) {
			cells = cells[:len(aligns)]
		}
		rowBlk := &Block{Kind: BlockTableRow}
		for _, c := range cells {
			rowBlk.Children = append(rowBlk.Children, &Block{Kind: BlockTableCell, Text: c})
		}
		table.Children = append(table.Children, rowBlk)
	}
	if i < len(lines) {
		leftover = strings.Join(lines[i:], "\n")
	}
	return table, leftover
}

// splitTableRow splits a pipe-delimited row, stripping outer pipes and
// honoring `\|` escapes.
func splitTableRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	var cells []string
	var b strings.Builder
	for i := 0; i < len(row); i++ {
		c := row[i]
		if c == '\\' && i+1 < len(row) && row[i+1] == '|' {
			b.WriteByte('|')
			i++
			continue
		}
		if c == '|' {
			cells = append(cells, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		}
		b.WriteByte(c)
	}
	cells = append(cells, strings.TrimSpace(b.String()))
	return cells
}

// parseSeparator parses a GFM table separator like "| :--- | ---: | :---: |"
// into a slice of Align values.
func parseSeparator(sep string) ([]Align, bool) {
	parts := splitTableRow(sep)
	if len(parts) == 0 {
		return nil, false
	}
	aligns := make([]Align, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			return nil, false
		}
		left := p[0] == ':'
		right := p[len(p)-1] == ':'
		body := p
		if left {
			body = body[1:]
		}
		if right && len(body) > 0 {
			body = body[:len(body)-1]
		}
		if body == "" {
			return nil, false
		}
		for i := 0; i < len(body); i++ {
			if body[i] != '-' {
				return nil, false
			}
		}
		switch {
		case left && right:
			aligns = append(aligns, AlignCenter)
		case left:
			aligns = append(aligns, AlignLeft)
		case right:
			aligns = append(aligns, AlignRight)
		default:
			aligns = append(aligns, AlignNone)
		}
	}
	return aligns, true
}
