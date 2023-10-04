FROM node:17-alpine AS angular-builder
WORKDIR /usr/src/app
ADD ui/package.json .
ADD ui/package-lock.json .
RUN npm install
COPY ui .
RUN npm run build

FROM golang:1.21 AS builder

COPY . /go/src/app
WORKDIR /go/src/app

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux make build

RUN git log -1 --oneline > version.txt

FROM alpine:latest

WORKDIR /root/

COPY --from=angular-builder /usr/src/app/dist ./ui/dist
COPY --from=builder /go/src/app/set_env.sh .
COPY --from=builder /go/src/app/analytics-cleanup .
COPY --from=builder /go/src/app/version.txt .

EXPOSE 8000

LABEL org.opencontainers.image.source https://github.com/SENERGY-Platform/analytics-cleanup

ENTRYPOINT ["sh", "set_env.sh"]