FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /ldap-manager ./cmd/ldap-manager

FROM scratch
COPY --from=builder /ldap-manager /ldap-manager
ENTRYPOINT ["/ldap-manager"]
