FROM golang:1.8

RUN mkdir -p /go/src/github.com/luizalabs/mitose
WORKDIR /go/src/github.com/luizalabs/mitose

ADD . /go/src/github.com/luizalabs/mitose
RUN go install

CMD ["mitose"]
