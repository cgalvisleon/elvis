ARG GO_VERSION=1.22.1

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk add --no-cache ca-certificates openssl git tzdata
RUN update-ca-certificates

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN gofmt -w . && go build ./cmd/apigateway

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/apigateway ./apigateway

ENTRYPOINT ["./apigateway"]
