FROM golang:1.22-alpine as builder

WORKDIR /app

COPY . .

RUN apk -U upgrade --ignore alpine-baselayout && apk add --no-cache \
    gcc \
    g++ \
    libxml2

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -a -installsuffix cgo -o ./cmd/main ./cmd


FROM alpine

RUN apk -U upgrade --ignore alpine-baselayout && apk add --no-cache \
    curl \
    nano \
    vim \
    bash \
    tzdata

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

RUN adduser -SDHh /app appuser

USER appuser

COPY --from=builder /app/cmd/main /app/bin/main
COPY --from=builder /app/.env /app/.env

WORKDIR /app/bin

CMD ["./main"]