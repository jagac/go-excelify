package converter

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/jagac/excelify/internal/types"
	"github.com/xuri/excelize/v2"
)

func adjustColumnWidths(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta) error {
	numCores := runtime.NumCPU()
	batchSize := (len(jsonData) + numCores - 1) / numCores

	colWidths := make(map[string]float64, len(meta))
	var mu sync.RWMutex
	var wg sync.WaitGroup

	for i := 0; i < len(jsonData); i += batchSize {
		end := i + batchSize
		if end > len(jsonData) {
			end = len(jsonData)
		}
		batch := jsonData[i:end]

		wg.Add(1)
		go func(batch []map[string]interface{}) {
			defer wg.Done()
			localColWidths := make(map[string]float64, len(meta))

			for _, row := range batch {
				for _, col := range meta {
					cellValue := fmt.Sprintf("%v", row[col.Name])
					width := float64(len(cellValue)) * 1.15
					if width < 10 {
						width = 10
					}
					if width > 255 {
						width = 255
					}
					if width > localColWidths[col.Name] {
						localColWidths[col.Name] = width
					}
				}
			}

			mu.Lock()
			for colName, width := range localColWidths {
				if width > colWidths[colName] {
					colWidths[colName] = width
				}
			}
			mu.Unlock()
		}(batch)
	}

	wg.Wait()

	for colName, width := range colWidths {
		colIndex := getColumnIndex(meta, colName)
		if colIndex != -1 {
			colStr := colIndexToName(colIndex)
			if err := f.SetColWidth(sheetName, colStr, colStr, width); err != nil {
				return err
			}
		}
	}

	return nil
}
