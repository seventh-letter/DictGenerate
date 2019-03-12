package table

import (
	"github.com/olekukonko/tablewriter"
	"io"
)

// PCSTable 封装 tablewriter.Table
type Table struct {
	*tablewriter.Table
}

// NewTable 预设了一些配置
func NewTable(wt io.Writer) Table {
	tb := tablewriter.NewWriter(wt)
	tb.SetAlignment(tablewriter.ALIGN_LEFT)
	tb.SetAutoWrapText(true)
	//tb.SetBorder(false)
	//tb.SetHeaderLine(false)
	//tb.SetColumnSeparator("")
	return Table{tb}
}
