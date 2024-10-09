package converter

import (
	"github.com/Jagac/excelify/types"
	"github.com/xuri/excelize/v2"
)

func createHeaders(meta []types.ColumnMeta) []string {
	var headers []string
	for _, col := range meta {
		headers = append(headers, col.Name)
	}

	return headers
}

func setHeaders(f *excelize.File, sheetName string, headers []string, style *ExcelStyles) error {
	for i, header := range headers {
		cell := colIndexToName(i) + "1"
		if err := f.SetCellStr(sheetName, cell, header); err != nil {
			return err
		}
	}
	if err := f.SetCellStyle(sheetName, "A1", colIndexToName(len(headers)-1)+"1", style.HeaderStyle); err != nil {
		return err
	}
	return nil
}
