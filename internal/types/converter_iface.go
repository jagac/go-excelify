package types

import (
	"bytes"

	"github.com/xuri/excelize/v2"
)


type Converter interface {
	ConvertToExcel(jsonData []map[string]interface{}, meta []ColumnMeta) (*bytes.Buffer, error)
	ConvertToJson(f *excelize.File) ([]byte, error)
}
