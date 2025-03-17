FROM golang:1.23.0-alpine

WORKDIR /app
COPY go.mod go.sum ./
COPY ./Task /app/Task
RUN go mod download
COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]