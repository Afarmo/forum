FROM golang:1.27.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -tags fts5 -o forum ./cmd

EXPOSE 8080

CMD ["./forum"]