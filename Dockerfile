FROM golang:1.22-alpine
WORKDIR /app
COPY . /app
RUN go build main.go
EXPOSE 8080
CMD ["./main"]
