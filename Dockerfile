# 建置階段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安裝 PostgreSQL 客戶端依賴
RUN apk add --no-cache postgresql-client

# 複製 go.mod 和 go.sum（如果有的話）
COPY go.mod ./
COPY go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 編譯應用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 執行階段
FROM alpine:latest

WORKDIR /app

# 安裝 PostgreSQL 客戶端依賴
RUN apk add --no-cache postgresql-client

# 從建置階段複製編譯好的執行檔
COPY --from=builder /app/main .

# 設定環境變數（建議使用環境變數而不是硬編碼的數據庫連接信息）
#ENV DB_HOST=lifeplan-instance-1.c1w4eg0mwku5.us-west-2.rds.amazonaws.com
#ENV DB_PORT=5432
#ENV DB_USER=username
#ENV DB_PASSWORD=password
#ENV DB_NAME=dbname

# 開放 8080 連接埠
EXPOSE 8080

# 執行應用
CMD ["./main"]