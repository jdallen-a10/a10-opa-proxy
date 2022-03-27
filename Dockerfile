##
## Dockerfile for a10-opa-proxy Container
##
## John D. Allen
## Global Solutions Architect - Cloud, IoT, & Automation
## A10 Networks Inc.
##
## March, 2022
##

FROM golang:1.17.3 AS builder

RUN mkdir /app
RUN mkdir /app/axapi
RUN mkdir /app/config
ADD ./axapi /app/axapi
ADD go.* /app
ADD opaproxy.go /app
ADD ./config/config.yaml /app/config
ADD Dockerfile /app

WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/opaproxy opaproxy.go

##
## Build Final Container
FROM alpine:latest AS production
COPY --from=builder /app .

CMD ["./opaproxy"]

