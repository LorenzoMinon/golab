FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o golab .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/golab .
COPY web/ ./web/
COPY projects/argodash/dashboard.html ./projects/argodash/dashboard.html

EXPOSE 8080
CMD ["./golab"]