FROM golang:1.15 as builder

WORKDIR /go/src/github.com/zsuzhengdu/grafana-annotations
COPY . /go/src/github.com/zsuzhengdu/grafana-annotations

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/grafana-annotations /go/src/github.com/zsuzhengdu/grafana-annotations/main.go

FROM alpine:3
RUN apk --update add ca-certificates
RUN addgroup -S grafana-annotations && adduser -S -G grafana-annotations grafana-annotations
USER grafana-annotations
COPY --from=builder /go/bin/grafana-annotations /usr/local/bin/grafana-annotations

ENTRYPOINT ["grafana-annotations"]
