FROM golang

RUN go get -u github.com/gobuffalo/packr/...
RUN mkdir /usr/local/packr
RUN cp /go/bin/packr /usr/local/go/bin

RUN apt-get update
RUN apt-get install -y zip

