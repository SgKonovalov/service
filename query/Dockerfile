FROM golang:latest
WORKDIR /show  
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build ./cmd/main
CMD ["./main"]