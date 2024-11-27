ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

ENV TEMPLATES_DIR=/templates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

FROM debian:bookworm

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/ 
COPY certs/prod-ca-2021.crt /usr/local/share/ca-certificates
RUN chmod 644 /usr/local/share/ca-certificates/prod-ca-2021.crt
RUN update-ca-certificates

COPY --from=builder /run-app /usr/local/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY templates /templates
COPY static /static
COPY data data

CMD ["run-app"]