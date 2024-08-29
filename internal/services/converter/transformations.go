package converter

import (
	"fmt"

	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/Jagac/excelify/types"
	"github.com/xuri/excelize/v2"
)

func createStyles(f *excelize.File) (*types.ExcelStyles, error) {
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Aptos Narrow",
			Bold:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	intStyle, err := f.NewStyle(&excelize.Style{NumFmt: 1})
	if err != nil {
		return nil, err
	}

	floatStyle, err := f.NewStyle(&excelize.Style{NumFmt: 2})
	if err != nil {
		return nil, err
	}

	exp := "yyyy-mm-dd"
	datetimeStyle, err := f.NewStyle(&excelize.Style{CustomNumFmt: &exp})
	if err != nil {
		return nil, err
	}

	percentageStyle, err := f.NewStyle(&excelize.Style{NumFmt: 10})
	if err != nil {
		return nil, err
	}

	textStyle, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return nil, err
	}

	hiddenFontColorStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: "#ff00ff",
		},
	})
	if err != nil {
		return nil, err
	}

	return &types.ExcelStyles{
		HeaderStyle:     headerStyle,
		IntStyle:        intStyle,
		FloatStyle:      floatStyle,
		DatetimeStyle:   datetimeStyle,
		PercentageStyle: percentageStyle,
		TextStyle:       textStyle,
		HiddenStyle:     hiddenFontColorStyle,
	}, nil
}

func createHeaders(meta []types.ColumnMeta) []string {
	var headers []string
	for _, col := range meta {
		headers = append(headers, col.Name)
	}

	return headers
}

func setData(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta, styles *types.ExcelStyles) error {
	const threshold = 10000

	if len(jsonData) <= threshold {
		return setDataSequential(f, sheetName, jsonData, meta, styles)
	}
	return setDataParallel(f, sheetName, jsonData, meta, styles)
}

func setDataSequential(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta, styles *types.ExcelStyles) error {
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

			var style int
			var err error

			switch colMeta.Type {
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
							return fmt.Errorf("failed to convert %v to integer: %w", strValue, err)
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
							return fmt.Errorf("failed to convert %v to float: %w", strValue, err)
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
							return fmt.Errorf("failed to convert %v to datetime: %w", strValue, err)
						}
					}
				}
			case "PERCENTAGE":
				style = styles.PercentageStyle
			default:
				continue
			}

			cellRef := colIndexToName(colIndex) + strconv.Itoa(rowIndex+2)
			if err := f.SetCellValue(sheetName, cellRef, value); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheetName, cellRef, cellRef, style); err != nil {
				return err
			}
		}
	}

	return nil
}

func setDataParallel(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta, styles *types.ExcelStyles) error {
	numCores := runtime.NumCPU()
	batchSize := len(jsonData) / numCores
	if len(jsonData)%numCores != 0 {
		batchSize++
	}

	cellDataChan := make(chan []types.CellData, numCores)
	errChan := make(chan error, numCores)
	var wg sync.WaitGroup

	processBatch := func(batch []map[string]interface{}, startIndex int) {
		defer wg.Done()
		var cellData []types.CellData

		for rowIndex, row := range batch {
			for colIndex, col := range meta {
				value := row[col.Name]

				var style int
				var err error

				switch col.Type {
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
								errChan <- fmt.Errorf("failed to convert %v to integer: %w", strValue, err)
								return
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
								errChan <- fmt.Errorf("failed to convert %v to float: %w", strValue, err)
								return
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
								errChan <- fmt.Errorf("failed to convert %v to datetime: %w", strValue, err)
								return
							}
						}
					}
				case "PERCENTAGE":
					style = styles.PercentageStyle
				default:
					continue
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

		batch := jsonData[i:end]
		wg.Add(1)
		go processBatch(batch, i)
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

func setColumnVisibility(f *excelize.File, sheetName string, meta []types.ColumnMeta, style *types.ExcelStyles) error {
	type task struct {
		colName string
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	numWorkers := runtime.NumCPU()
	taskChan := make(chan task, len(meta))
	worker := func() {
		for t := range taskChan {
			colName := t.colName
			for rowIndex := range rows {
				cell := fmt.Sprintf("%s%d", colName, rowIndex+1)
				if err := f.SetCellStyle(sheetName, cell, cell, style.HiddenStyle); err != nil {
					continue
				}
			}
		}
	}

	var wg sync.WaitGroup
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
				continue
			}
			taskChan <- task{colName: colName}
		}
	}

	close(taskChan)
	wg.Wait()

	return nil
}

func colIndexToName(index int) string {
	var columnName string
	for index >= 0 {
		columnName = string(rune('A'+index%26)) + columnName
		index = index/26 - 1
	}
	return columnName
}

func adjustColumnWidths(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta) error {
	numCores := runtime.NumCPU()
	batchSize := len(jsonData) / numCores

	colWidths := make(map[string]float64)
	var mu sync.Mutex
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
			localColWidths := make(map[string]float64)

			for _, row := range batch {
				for _, col := range meta {
					cellValue := fmt.Sprintf("%v", row[col.Name])
					colName := col.Name

					width := float64(len(cellValue)) * 1.15
					if width < 10 {
						width = 10
					}
					if width > 255 {
						width = 255
					}
					if width > localColWidths[colName] {
						localColWidths[colName] = width
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

func setHeaders(f *excelize.File, sheetName string, headers []string, style *types.ExcelStyles) error {
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

func getColumnIndex(meta []types.ColumnMeta, columnName string) int {
	for i, col := range meta {
		if col.Name == columnName {
			return i
		}
	}
	return -1
}
