# 1. Image
FROM golang:1.24-alpine AS builder

# 2. Working directory
WORKDIR /app

# 3. Get required dependencies
# 3. Копируем go.mod и go.sum из корня (2 уровня вверх)
COPY go.mod go.sum ./

# 4. Загрузка зависимостей

RUN go mod download
# 4. Copy src code
COPY . .

# 5. Assembly app
RUN go build -o producer ./main.go

# 6. Minmial workin assembly
FROM alpine:latest

WORKDIR /root/

# Copy binary file from previous builder

COPY --from=builder /app/producer .
COPY --from=builder /app/temp/media media/

CMD ["./producer"]
LABEL authors="officialevsty"
