FROM golang:1.19-buster as builder

WORKDIR /dist

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN make build

FROM golang:1.19-buster
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chown=0:0 --from=builder /dist/bin/app /dist/app

WORKDIR /dist
ENTRYPOINT ["/dist/app"]