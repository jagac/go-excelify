package converter

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/jagac/excelify/types"
	"github.com/xuri/excelize/v2"
)

func setColumnVisibility(f *excelize.File, sheetName string, meta []types.ColumnMeta, style *ExcelStyles) error {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	numWorkers := runtime.NumCPU()
	taskChan := make(chan string, len(meta))

	var wg sync.WaitGroup

	worker := func() {
		for colName := range taskChan {
			for rowIndex := range rows {
				cell := fmt.Sprintf("%s%d", colName, rowIndex+1)
				_ = f.SetCellStyle(sheetName, cell, cell, style.HiddenStyle)
			}
		}
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			worker()
		}()
	}

	for colIndex, col := range meta {
		if col.DefaultVisibility == "hidden" || col.DefaultVisibility == "always_hidden" {
			colName := colIndexToName(colIndex)
			if err := f.SetColVisible(sheetName, colName, false); err != nil {
				return err
			}
			taskChan <- colName
		}
	}

	close(taskChan)
	wg.Wait()

	return nil
}
