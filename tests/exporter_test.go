package tests

import (
	"bytes"
	"encoding/json"
	"fmt"

	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Jagac/excelify/internal/server"
	"github.com/Jagac/excelify/internal/services/converter"
	"github.com/Jagac/excelify/internal/services/logging"
	"github.com/Jagac/excelify/types"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

func TestJsonHandler(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	router := http.NewServeMux()
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}

	converter := converter.NewConverter()
	handler := server.NewHandler(converter, logger)
	handler.RegisterRoutes(router)

	t.Run("should convert using sequential", func(t *testing.T) {

		payload := types.RequestJson{
			Filename: "example.xlsx",
			Data:     generateDataItems(1000),
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

		marshalled, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/api/v1/conversions", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := http.NewServeMux()

		router.HandleFunc("POST /api/v1/conversions", handler.HandleJsonToExcel)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		responseBody := rr.Body.Bytes()
		file, err := os.Create("response.xlsx")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		_, err = file.Write(responseBody)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat("response.xlsx"); os.IsNotExist(err) {
			t.Fatal("file was not created")
		}
		fmt.Println("Excel file saved successfully as response.xlsx")
		checkExcelColumnsAndData(t, "response.xlsx", []string{"name", "age", "email", "salary", "joined"})
		e := os.Remove("response.xlsx")
		if e != nil {
			log.Fatal(e)
		}
	})

	t.Run("should convert using parallel", func(t *testing.T) {

		payload := types.RequestJson{
			Filename: "example.xlsx",
			Data:     generateDataItems(50000),
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

		marshalled, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/api/v1/conversions", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := http.NewServeMux()

		router.HandleFunc("POST /api/v1/conversions", handler.HandleJsonToExcel)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		responseBody := rr.Body.Bytes()
		file, err := os.Create("response.xlsx")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		_, err = file.Write(responseBody)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat("response.xlsx"); os.IsNotExist(err) {
			t.Fatal("file was not created")
		}
		fmt.Println("Excel file saved successfully as response.xlsx")
	})
	checkExcelColumnsAndData(t, "response.xlsx", []string{"name", "age", "email", "salary", "joined"})
	e := os.Remove("response.xlsx")
	if e != nil {
		log.Fatal(e)
	}
	os.RemoveAll("logs")

}

func generateDataItems(n int) []map[string]interface{} {
	data := make([]map[string]interface{}, n)
	for i := 0; i < n; i++ {
		data[i] = map[string]interface{}{
			"name":   fmt.Sprintf("Name %d", i),
			"age":    20 + (i % 50),
			"email":  fmt.Sprintf("email%d@example.com", i),
			"salary": 30000 + (float64(i) * 10),
			"joined": fmt.Sprintf("2022-01-%02d 15:04", (i%31)+1),
		}
	}
	return data
}

func checkExcelColumnsAndData(t *testing.T, filePath string, expectedColumns []string) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		t.Fatalf("failed to open Excel file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close Excel file: %v", err)
		}
	}()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatalf("failed to get rows from Excel file: %v", err)
	}

	if len(rows) == 0 {
		t.Fatal("no rows found in the Excel file")
	}

	headerRow := rows[0]
	if len(headerRow) != len(expectedColumns) {
		t.Fatalf("expected %d columns, got %d", len(expectedColumns), len(headerRow))
	}

	for i, expectedColumn := range expectedColumns {
		if headerRow[i] != expectedColumn {
			t.Errorf("expected column %s at index %d, got %s", expectedColumn, i, headerRow[i])
		}
	}

	nameIndex := -1
	for i, colName := range headerRow {
		if colName == "name" {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		t.Fatal("no 'name' column found in the headers")
	}

	for rowIndex, row := range rows[1:] {
		if len(row) <= nameIndex {
			t.Errorf("row %d is missing the 'name' column", rowIndex+2)
			continue
		}

		expectedName := fmt.Sprintf("Name %d", rowIndex)
		if row[nameIndex] != expectedName {
			t.Errorf("expected name '%s' in row %d, got '%s'", expectedName, rowIndex+2, row[nameIndex])
		}
	}
}
