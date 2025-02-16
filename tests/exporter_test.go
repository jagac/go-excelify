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

	"github.com/jagac/excelify/internal/converter"
	"github.com/jagac/excelify/internal/logging"
	"github.com/jagac/excelify/internal/server"
	"github.com/jagac/excelify/internal/types"
	"github.com/joho/godotenv"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {

	goleak.VerifyTestMain(m)
}

func TestJsonHandler(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	mux := http.NewServeMux()
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}
	converter := converter.NewConverter()

	handler := server.NewHandler(converter)
	router := server.NewRouter(handler, logger)
	router.RegisterRoutes(mux)

	t.Run("should convert using sequential", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		payload := types.RequestJson{
			Filename: "example.xlsx",
			Data:     GenerateDataItems(1000),
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
		CheckExcelColumnsAndData(t, "response.xlsx", []string{"name", "age", "email", "salary", "joined"})
		e := os.Remove("response.xlsx")
		if e != nil {
			log.Fatal(e)
		}
	})

	t.Run("should convert using parallel", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		payload := types.RequestJson{
			Filename: "example.xlsx",
			Data:     GenerateDataItems(50000),
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
	CheckExcelColumnsAndData(t, "response.xlsx", []string{"name", "age", "email", "salary", "joined"})
	e := os.Remove("response.xlsx")
	if e != nil {
		log.Fatal(e)
	}
	os.RemoveAll("logs")

}
