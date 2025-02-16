package tests

import (
	"testing"

	"github.com/jagac/excelify/internal/converter"
	"github.com/jagac/excelify/internal/types"
)

func BenchmarkConversion(b *testing.B) {

	conv := converter.NewConverter()
	payload := types.RequestJson{
		Filename: "example.xlsx",
		Data:     GenerateDataItems(100000),
		Meta: types.MetaData{
			Columns: []types.ColumnMeta{
				{
					Name:              "name",
					Type:              "STRING",
					DefaultVisibility: "hidden",
				},
				{
					Name: "age",
					Type: "INTEGER",
				},
				{
					Name: "email",
					Type: "STRING",
				},
				{
					Name: "salary",
					Type: "FLOAT",
				},
				{
					Name: "joined",
					Type: "DATETIME",
				},
			},
		},
	}

	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertToExcel(payload.Data, payload.Meta.Columns)
		if err != nil {
			b.Fatalf("Error occurred during ConvertToJson: %v", err)
		}
	}
}
