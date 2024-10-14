package converter

import (
	"fmt"

	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/jagac/excelify/types"
	"github.com/xuri/excelize/v2"
)

func setData(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta, styles *ExcelStyles) error {
	const threshold = 10000

	if len(jsonData) <= threshold {
		return setDataSequential(f, sheetName, jsonData, meta, styles)
	}
	return setDataParallel(f, sheetName, jsonData, meta, styles)
}

func setDataSequential(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta, styles *ExcelStyles) error {
	metaIndex := make(map[string]int)
	for i, col := range meta {
		metaIndex[col.Name] = i
	}

	for rowIndex, row := range jsonData {
		orderedRow := make([]interface{}, len(meta))

		for colName, value := range row {
			if colIndex, exists := metaIndex[colName]; exists {
				orderedRow[colIndex] = value
			}
		}

		for colIndex, value := range orderedRow {
			colMeta := meta[colIndex]

			convertedValue, style, err := convertValue(value, colMeta.Type, styles)
			if err != nil {
				return err
			}

			cellRef := colIndexToName(colIndex) + strconv.Itoa(rowIndex+2)
			if err := f.SetCellValue(sheetName, cellRef, convertedValue); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheetName, cellRef, cellRef, style); err != nil {
				return err
			}
		}
	}

	return nil
}

func setDataParallel(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta, styles *ExcelStyles) error {
	numCores := runtime.NumCPU()
	batchSize := (len(jsonData) + numCores - 1) / numCores

	cellDataChan := make(chan []types.CellData, numCores)
	errChan := make(chan error, numCores)
	var wg sync.WaitGroup

	processBatch := func(batch []map[string]interface{}, startIndex int) {
		defer wg.Done()
		var cellData []types.CellData

		for rowIndex, row := range batch {
			for colIndex, col := range meta {
				value, style, err := convertValue(row[col.Name], col.Type, styles)
				if err != nil {
					errChan <- err
					return
				}

				cellData = append(cellData, types.CellData{
					RowIndex: startIndex + rowIndex,
					ColIndex: colIndex,
					Value:    value,
					Style:    style,
				})
			}
		}

		cellDataChan <- cellData
	}

	for i := 0; i < len(jsonData); i += batchSize {
		end := i + batchSize
		if end > len(jsonData) {
			end = len(jsonData)
		}

		wg.Add(1)
		go processBatch(jsonData[i:end], i)
	}

	go func() {
		wg.Wait()
		close(cellDataChan)
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	for cellData := range cellDataChan {
		for _, cell := range cellData {
			cellRef := colIndexToName(cell.ColIndex) + strconv.Itoa(cell.RowIndex+2)
			if err := f.SetCellValue(sheetName, cellRef, cell.Value); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheetName, cellRef, cellRef, cell.Style); err != nil {
				return err
			}
		}
	}

	return nil
}

func convertValue(value interface{}, colType string, styles *ExcelStyles) (interface{}, int, error) {
	var style int
	var err error

	switch colType {
	case "STRING":
		style = styles.TextStyle
		if value == nil {
			value = ""
		}
	case "INTEGER":
		style = styles.IntStyle
		if strValue, ok := value.(string); ok {
			if strValue == "" {
				value = ""
			} else {
				value, err = strconv.Atoi(strValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to convert %v to integer: %w", strValue, err)
				}
			}
		}
	case "FLOAT":
		style = styles.FloatStyle
		if strValue, ok := value.(string); ok {
			if strValue == "" {
				value = ""
			} else {
				value, err = strconv.ParseFloat(strValue, 64)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to convert %v to float: %w", strValue, err)
				}
			}
		}
	case "DATETIME":
		style = styles.DatetimeStyle
		if strValue, ok := value.(string); ok {
			if strValue == "" {
				value = ""
			} else {
				value, err = time.Parse("2006-01-02 15:04", strValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to convert %v to datetime: %w", strValue, err)
				}
			}
		}
	case "PERCENTAGE":
		style = styles.PercentageStyle
	default:
		return nil, 0, nil
	}

	return value, style, nil
}
