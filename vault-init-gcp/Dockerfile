FROM golang:1.12.7 as builder

# Install deps
WORKDIR /workspace
COPY go.* .
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o vault-init-gcp

# Use distroless as minimal base image to package the vault-init-gcp binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/vault-init-gcp .
ENTRYPOINT ["/vault-init-gcp"]
