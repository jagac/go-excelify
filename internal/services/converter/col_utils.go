package converter

import "github.com/jagac/excelify/types"

func colIndexToName(index int) string {
	var columnName string
	for index >= 0 {
		columnName = string(rune('A'+index%26)) + columnName
		index = index/26 - 1
	}
	return columnName
}

func getColumnIndex(meta []types.ColumnMeta, columnName string) int {
	for i, col := range meta {
		if col.Name == columnName {
			return i
		}
	}
	return -1
}
