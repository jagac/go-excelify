package converter

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Jagac/excelify/types"
	"github.com/xuri/excelize/v2"
)

type ConverterImpl struct{}

func NewConverter() types.Converter {
	return &ConverterImpl{}
}

func (c *ConverterImpl) ConvertToExcel(jsonData []map[string]interface{}, meta []types.ColumnMeta) (*bytes.Buffer, error) {

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

func (c *ConverterImpl) ConvertToJson(f *excelize.File) ([]byte, error) {
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	var result []map[string]interface{}

	headers := rows[0]
	for _, row := range rows[1:] {
		rowData := make(map[string]interface{})
		for i, cell := range row {
			rowData[headers[i]] = cell
		}
		result = append(result, rowData)
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, nil
}
