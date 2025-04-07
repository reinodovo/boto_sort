# syntax=docker/dockerfile:1
FROM golang:1.24-alpine
WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN go build ./cmd/boto_sort

CMD ["./boto_sort"]
