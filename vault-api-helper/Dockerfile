FROM golang:1.12.7 as builder

RUN mkdir /go/src/app
WORKDIR /go/src/app
ENV GOPATH=/go
ENV GO111MODULE=on
RUN go mod init
RUN go get k8s.io/client-go@v12.0.0
ADD ./main.go /go/src/app

#COPY Gopkg.toml /go/src/app

RUN go test -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app

# Use distroless as minimal base image to package the vault-api-helper binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:latest
#WORKDIR /
#COPY --from=builder /go/src/app/app .
#ENTRYPOINT ["/app"]

FROM alpine:3.9
RUN apk add --no-cache ca-certificates
CMD ["./app"]
COPY --from=builder /go/src/app/app .
