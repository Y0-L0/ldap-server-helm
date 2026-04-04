FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go test -c -o /e2e.test ./test/e2e/

FROM artifacts.software-univention.de/nubus/images/ldap-server:0.47.13 AS e2e
COPY --from=e2e-builder /e2e.test /e2e/e2e.test
COPY test/e2e/testdata/ /e2e/testdata/
WORKDIR /e2e
ENV INTEGRATION=1
ENTRYPOINT ["./e2e.test", "-test.v"]
