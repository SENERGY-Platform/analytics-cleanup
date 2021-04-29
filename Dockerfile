FROM golang:1.16 AS builder

COPY . /go/src/app
WORKDIR /go/src/app

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux make build

RUN git log -1 --oneline > version.txt

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/app/analytics-cleanup .
COPY --from=builder /go/src/app/version.txt .

ENTRYPOINT ["./analytics-cleanup"]