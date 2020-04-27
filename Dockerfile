FROM golang:1.11 as builder
WORKDIR /go/src/github.com/luizalabs/mitose
ADD . /go/src/github.com/luizalabs/mitose
RUN CGO_ENABLED=0 go build -o mitose

FROM alpine:3.7
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /go/src/github.com/luizalabs/mitose/mitose .
CMD ["./mitose"]
