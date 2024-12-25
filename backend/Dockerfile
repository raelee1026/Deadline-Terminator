FROM golang:1.23.0-alpine

# 設定工作目錄
WORKDIR /app
COPY go.mod go.sum ./
COPY ./Task /app/Task
RUN go mod download
COPY . .
#COPY ../config /app/config
# 安裝依賴並構建
RUN go build -o main .

# 曝露服務埠
EXPOSE 8080

# 啟動後端服務
CMD ["./main"]