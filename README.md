# Project go-excelify
A stateless microservice that converts json to excel. Useful if you need to export some data to excel for a user.

# Docs

### Expected body schema: <br>
Note: Expected types in meta need to be specified as they are required for excel styling

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
    },
    {
      "name": "Jane Smith",
      "age": 25,
      "email": "jane.smith@example.com",
      "salary": 48000.75,
      "joined": "2021-07-23 15:04"
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

### Docker
```bash
docker build -t excelify -f .dockerbuild/Dockerfile .
```


### MakeFile
run the application

```bash
make run
```

run all make commands with tests

```bash
make all
```

build the application

```bash
make build
```

run the test suite

```bash
make test
```

run the linter

```bash
make lint
```

run security scan

```bash
make scan
```
