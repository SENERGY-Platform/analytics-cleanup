FROM golang:1.14

COPY . /go/src/analytics-cleanup
WORKDIR /go/src/analytics-cleanup

RUN make build

CMD ./analytics-cleanup