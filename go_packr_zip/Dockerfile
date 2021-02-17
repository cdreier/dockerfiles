FROM golang:1.16.0

RUN go get -u github.com/gobuffalo/packr/v2/packr2

RUN go get -u github.com/markbates/pkger/cmd/pkger

RUN cp /go/bin/packr2 /usr/local/go/bin
RUN cp /go/bin/pkger /usr/local/go/bin

RUN apt-get update
RUN apt-get install -y zip
RUN apt-get install -y gcc-multilib
RUN apt-get install -y gcc-mingw-w64
