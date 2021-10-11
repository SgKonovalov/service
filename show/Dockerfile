FROM golang:latest
WORKDIR /show  
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build ./cmd/web
CMD ["./web"]
EXPOSE 8080