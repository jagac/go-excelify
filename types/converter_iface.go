package types

import "bytes"


type Converter interface {
	ConvertToExcel(jsonData []map[string]interface{}, meta []ColumnMeta) (*bytes.Buffer, error)
}