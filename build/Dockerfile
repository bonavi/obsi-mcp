# syntax=docker/dockerfile:1
FROM golang:1.26 AS builder

ARG VERSION
ARG BUILD
ARG BUILD_DATE
ADD . /obsi-mcp
WORKDIR /obsi-mcp/server
RUN sed -i -e "s|@{build}|${BUILD}|g" -e "s|@{version}|${VERSION}|g" -e "s|@{buildDate}|${BUILD_DATE}|g" internal/main.go
RUN CGO_ENABLED=0 go build -trimpath -o ../main internal/main.go

FROM alpine:3.22
# tini — PID 1: забирает осиротевших потомков. git отсоединяет фоновые gc/maintenance,
# а при таймауте CommandContext остаётся сирота ssh; ./main как PID 1 их не реапит,
# и на ноде копились тысячи зомби (2026-11-11)
RUN apk add --no-cache git openssh-client tini && adduser -D -u 1000 mcp
WORKDIR /obsi-mcp
COPY --from=builder /obsi-mcp/main .
EXPOSE 8080
EXPOSE 8090
USER 1000
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./main"]
