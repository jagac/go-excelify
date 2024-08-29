# Project go-excelify
A stateless microservice that converts json to excel and excel to json. <br>
Useful if your app needs a quick way to export some data and give it to a user as excel (business people love excel) or get some data converted from excel so your other service can consume it.


---

## Table of Contents

- [Docs](#docs)

  - [Docker](#docker-15-mb)
  - [API Endpoints](#api-endpoints)
    - [Convert JSON to Excel](#convert-json-to-excel)
    - [Convert Excel to JSON](#convert-excel-to-json)


---

# Docs


## Docker (~15 MB)
```yaml
services:
  excelify:
    build:
      context: .
      dockerfile: .dockerbuild/Dockerfile
    ports:
      - "${PORT}:${PORT}"
    volumes:
      - ./logs:/root/logs
    env_file:
      - .env
    restart: always
```
Note: you need so specify LOG_DIR and PORT in a .env for the build.
I suggest:
`LOG_DIR=/root/logs
PORT=3000`



---

## API Endpoints

### Convert JSON to Excel

- **Endpoint:** `POST /api/v1/conversions/to-excel`
- **Description:** Converts JSON data into an Excel file.
- **Headers:**
  - `Content-Type: application/json`
- **Request Body:**
  - The request body should contain a JSON object with the following structure:

    ```json
    {
      "filename": "example.xlsx",
      "data": [
        {
          "name": "John Doe",
          "age": 30,
          "email": "john.doe@example.com",
          "salary": 55000.5,
          "joined": "2022-01-15 15:04"
        }
      ],
      "meta": {
        "columns": [
          {
            "name": "name",
            "type": "STRING",
            "default_visibility": "hidden"
          },
          {
            "name": "age",
            "type": "INTEGER"
          },
          {
            "name": "email",
            "type": "STRING"
          },
          {
            "name": "salary",
            "type": "FLOAT"
          },
          {
            "name": "joined",
            "type": "DATETIME"
          }
        ]
      }
    }
    ```

- **Response:**
  - **Success:**
    - **Status:** `200 OK`
    - **Headers:**
      - `Content-Disposition: attachment; filename=example.xlsx`
      - `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet; charset=utf-8`
    - **Body:** Returns the generated Excel file.
  - **Error:**
    - **Status:** `400 Bad Request` if the JSON is malformed or no data is provided.
    - **Status:** `500 Internal Server Error` if there is an issue with the conversion process.

- **Example Request:**

    ```bash
    curl -X POST https://yourdomain.com/api/v1/conversions/to-excel \
      -H "Content-Type: application/json" \
      -d '{
            "filename": "example.xlsx",
            "data": [
              {
                "name": "John Doe",
                "age": 30,
                "email": "john.doe@example.com",
                "salary": 55000.5,
                "joined": "2022-01-15 15:04"
              }
            ],
            "meta": {
              "columns": [
                {
                  "name": "name",
                  "type": "STRING",
                  "default_visibility": "hidden"
                },
                {
                  "name": "age",
                  "type": "INTEGER"
                },
                {
                  "name": "email",
                  "type": "STRING"
                },
                {
                  "name": "salary",
                  "type": "FLOAT"
                },
                {
                  "name": "joined",
                  "type": "DATETIME"
                }
              ]
            }
          }'
    ```

---

### Convert Excel to JSON

- **Endpoint:** `POST /api/v1/conversions/to-json`
- **Description:** Converts an Excel file into JSON data.
- **Headers:**
  - `Content-Type: multipart/form-data`
- **Request Body:**
  - The request body should contain the Excel file as form data with the key `file`.

- **Response:**
  - **Success:**
    - **Status:** `200 OK`
    - **Headers:**
      - `Content-Type: application/json`
    - **Body:** Returns the JSON representation of the Excel file.
  - **Error:**
    - **Status:** `400 Bad Request` if the file cannot be read or parsed.
    - **Status:** `500 Internal Server Error` if there is an issue with the conversion process.

- **Example Request:**

    ```bash
    curl -X POST https://yourdomain.com/api/v1/conversions/to-json \
      -H "Content-Type: multipart/form-data" \
      -F "file=@/path/to/yourfile.xlsx"
    ```

---
