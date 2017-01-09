FROM golang:1.8-alpine

RUN apk add --no-cache git

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
CMD ["go-wrapper", "run"]
COPY main.go /go/src/app
RUN go-wrapper download
RUN go-wrapper install