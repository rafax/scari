FROM golang:1.8beta2-alpine

RUN apk add --no-cache youtube-dl ffmpeg git

RUN mkdir -p /go/src/app
RUN mkdir /out
WORKDIR /go/src/app
CMD ["go-wrapper", "run"]
COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
