FROM golang

RUN go get -u github.com/gobuffalo/packr/...
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN cp /go/bin/packr /usr/local/go/bin
RUN cp /go/bin/packr2 /usr/local/go/bin

RUN apt-get update
RUN apt-get install -y zip
