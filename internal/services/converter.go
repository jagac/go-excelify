package services

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/Jagac/excelify/types"

	"github.com/xuri/excelize/v2"
)


type ToExcelConverter struct{}

func NewConverter() types.Converter {
	return &ToExcelConverter{}
}

func (c *ToExcelConverter) ConvertToExcel(jsonData []map[string]interface{}, meta []types.ColumnMeta) (*bytes.Buffer, error) {

	f := excelize.NewFile()
	sheetName := "Sheet1"

	if _, err := f.NewSheet(sheetName); err != nil {
		return nil, err
	}

	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, err
	}

	styles, err := createStyles(f)
	if err != nil {
		return nil, err
	}

	headers := createHeaders(meta)
	if err := setHeaders(f, sheetName, headers, styles); err != nil {
		return nil, err
	}

	if err := setData(f, sheetName, jsonData, meta, styles); err != nil {
		return nil, err
	}

	if err := adjustColumnWidths(f, sheetName, jsonData, meta); err != nil {
		return nil, err
	}

	lastColIndex := len(meta) - 1
	lastColName := colIndexToName(lastColIndex)
	rangeString := "A1:" + lastColName + "1"
	if err := f.AutoFilter(sheetName, rangeString, []excelize.AutoFilterOptions{}); err != nil {
		return nil, err
	}

	if err := setColumnVisibility(f, sheetName, meta, styles); err != nil {
		return nil, err
	}

	if err := f.SetDefaultFont("Aptos Narrow"); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return nil, err
	}

	return &buffer, nil
}

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

func setHeaders(f *excelize.File, sheetName string, headers []string,  style *types.ExcelStyles) error {
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
	batchSize := (len(jsonData) + numCores - 1) / numCores

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	cellDataChan := make(chan []types.CellData, numCores)

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
								mu.Lock()
								if firstErr == nil {
									firstErr = fmt.Errorf("failed to convert %v to integer: %w", strValue, err)
								}
								mu.Unlock()
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
								mu.Lock()
								if firstErr == nil {
									firstErr = fmt.Errorf("failed to convert %v to float: %w", strValue, err)
								}
								mu.Unlock()
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
								mu.Lock()
								if firstErr == nil {
									firstErr = fmt.Errorf("failed to convert %v to datetime: %w", strValue, err)
								}
								mu.Unlock()
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

		wg.Add(1)
		go processBatch(jsonData[i:end], i)
	}

	go func() {
		wg.Wait()
		close(cellDataChan)
	}()

	var allCellData []types.CellData
	for cellData := range cellDataChan {
		allCellData = append(allCellData, cellData...)
	}

	if firstErr != nil {
		return firstErr
	}

	for _, cell := range allCellData {
		cellRef := colIndexToName(cell.ColIndex) + strconv.Itoa(cell.RowIndex+2)
		if err := f.SetCellValue(sheetName, cellRef, cell.Value); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, cellRef, cellRef, cell.Style); err != nil {
			return err
		}
	}

	return nil
}

func setColumnVisibility(f *excelize.File, sheetName string, meta []types.ColumnMeta, styles *types.ExcelStyles) error {
	type task struct {
		colName string
	}

	numWorkers := runtime.NumCPU()
	taskChan := make(chan task, len(meta))

	worker := func() {
		for t := range taskChan {
			colName := t.colName
			if err := f.SetColVisible(sheetName, colName, false); err != nil {
				continue
			}
			rows, err := f.GetRows(sheetName)
			if err != nil {
				continue
			}
			for rowIndex := range rows {
				cell := fmt.Sprintf("%s%d", colName, rowIndex+1)
				if err := f.SetCellStyle(sheetName, cell, cell, styles.HiddenStyle); err != nil {
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
			taskChan <- task{colName: colName}
		}
	}

	close(taskChan)
	wg.Wait()

	return nil
}

func adjustColumnWidths(f *excelize.File, sheetName string, jsonData []map[string]interface{}, meta []types.ColumnMeta) error {
	numCores := runtime.NumCPU()
	batchSize := (len(jsonData) + numCores - 1) / numCores

	colWidths := make([]map[string]float64, numCores)
	for i := range colWidths {
		colWidths[i] = make(map[string]float64)
	}

	var wg sync.WaitGroup
	work := make(chan int, numCores)

	worker := func(workerID int) {
		defer wg.Done()
		for batchStart := range work {
			localColWidths := colWidths[workerID]
			batchEnd := batchStart + batchSize
			if batchEnd > len(jsonData) {
				batchEnd = len(jsonData)
			}

			for _, row := range jsonData[batchStart:batchEnd] {
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
		}
	}

	for i := 0; i < numCores; i++ {
		wg.Add(1)
		go worker(i)
	}

	for i := 0; i < len(jsonData); i += batchSize {
		work <- i
	}
	close(work)
	wg.Wait()

	finalColWidths := make(map[string]float64)
	for _, widths := range colWidths {
		for colName, width := range widths {
			if finalColWidths[colName] < width {
				finalColWidths[colName] = width
			}
		}
	}

	for colName, width := range finalColWidths {
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



func getColumnIndex(meta []types.ColumnMeta, columnName string) int {
	for i, col := range meta {
		if col.Name == columnName {
			return i
		}
	}
	return -1
}

func colIndexToName(index int) string {
	var columnName string
	for index >= 0 {
		columnName = string(rune('A'+index%26)) + columnName
		index = index/26 - 1
	}
	return columnName
}
