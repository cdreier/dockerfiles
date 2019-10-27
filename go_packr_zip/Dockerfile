FROM golang:1.13.3

RUN go get -u github.com/gobuffalo/packr/...
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN go get -u github.com/inconshreveable/mousetrap
RUN GOOS=windows go get github.com/konsorten/go-windows-terminal-sequences

RUN mkdir --parents /usr/local/go/src/github.com/inconshreveable
RUN mv /go/src/github.com/inconshreveable/mousetrap /usr/local/go/src/github.com/inconshreveable

RUN mkdir --parents /usr/local/go/src/github.com/konsorten
RUN mv /go/src/github.com/konsorten/go-windows-terminal-sequences /usr/local/go/src/github.com/konsorten

RUN cp /go/bin/packr /usr/local/go/bin
RUN cp /go/bin/packr2 /usr/local/go/bin

RUN apt-get update
RUN apt-get install -y zip
