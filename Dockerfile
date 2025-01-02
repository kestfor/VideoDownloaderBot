FROM golang:1.23.1-alpine

WORKDIR /videoDownloader

COPY code/ .

RUN go mod download

RUN go build -o main .

CMD ["/videoDownloader/main"]