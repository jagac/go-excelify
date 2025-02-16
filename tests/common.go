package tests

import (
	"fmt"
	"testing"

	"github.com/xuri/excelize/v2"
)


func GenerateDataItems(n int) []map[string]interface{} {
	data := make([]map[string]interface{}, n)
	for i := 0; i < n; i++ {
		data[i] = map[string]interface{}{
			"name":   fmt.Sprintf("Name %d", i),
			"age":    20 + (i % 50),
			"email":  fmt.Sprintf("email%d@example.com", i),
			"salary": 30000 + (float64(i) * 10),
			"joined": fmt.Sprintf("2022-01-%02d 15:04", (i%31)+1),
		}
	}
	return data
}


func CheckExcelColumnsAndData(t *testing.T, filePath string, expectedColumns []string) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		t.Fatalf("failed to open Excel file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close Excel file: %v", err)
		}
	}()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatalf("failed to get rows from Excel file: %v", err)
	}

	if len(rows) == 0 {
		t.Fatal("no rows found in the Excel file")
	}

	headerRow := rows[0]
	if len(headerRow) != len(expectedColumns) {
		t.Fatalf("expected %d columns, got %d", len(expectedColumns), len(headerRow))
	}

	for i, expectedColumn := range expectedColumns {
		if headerRow[i] != expectedColumn {
			t.Errorf("expected column %s at index %d, got %s", expectedColumn, i, headerRow[i])
		}
	}

	nameIndex := -1
	for i, colName := range headerRow {
		if colName == "name" {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		t.Fatal("no 'name' column found in the headers")
	}

	for rowIndex, row := range rows[1:] {
		if len(row) <= nameIndex {
			t.Errorf("row %d is missing the 'name' column", rowIndex+2)
			continue
		}

		expectedName := fmt.Sprintf("Name %d", rowIndex)
		if row[nameIndex] != expectedName {
			t.Errorf("expected name '%s' in row %d, got '%s'", expectedName, rowIndex+2, row[nameIndex])
		}
	}
}