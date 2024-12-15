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

	// Initial width setting per column with 10 as the minimum width
	globalColWidths := make(map[string]float64, len(meta))
	for _, col := range meta {
		globalColWidths[col.Name] = 10.0
	}

	var wg sync.WaitGroup
	localWidthsChan := make(chan map[string]float64, numCores)

	// Process each batch in a goroutine
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
					cellValue, exists := row[col.Name]
					if !exists {
						continue
					}

					width := float64(len(fmt.Sprintf("%v", cellValue))) * 1.15
					if width < 10 {
						width = 10
					} else if width > 255 {
						width = 255
					}

					if width > localColWidths[col.Name] {
						localColWidths[col.Name] = width
					}
				}
			}

			localWidthsChan <- localColWidths
		}(batch)
	}

	// Close channel after all goroutines complete
	go func() {
		wg.Wait()
		close(localWidthsChan)
	}()

	// Merge local column widths into globalColWidths
	for localWidths := range localWidthsChan {
		for colName, width := range localWidths {
			if width > globalColWidths[colName] {
				globalColWidths[colName] = width
			}
		}
	}

	// Apply column widths to the Excel sheet
	for colName, width := range globalColWidths {
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
