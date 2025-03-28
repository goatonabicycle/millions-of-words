ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

ENV TEMPLATES_DIR=/templates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

FROM debian:bookworm

RUN apt-get update && apt-get install -y ca-certificates
RUN update-ca-certificates

COPY --from=builder /run-app /usr/local/bin/
COPY templates /templates
COPY static /static

CMD ["run-app"]