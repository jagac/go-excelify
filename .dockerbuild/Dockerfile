FROM golang:1.22.4-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /go/bin/api cmd/api/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /root/

COPY --from=builder /go/bin/api .
COPY .env .
EXPOSE ${PORT}
CMD ["./api"]
