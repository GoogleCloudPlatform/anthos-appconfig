FROM golang:1.12.7 as builder

RUN mkdir /go/src/app
WORKDIR /go/src/app
ENV GOPATH=/go
ENV GO111MODULE=on
RUN go mod init
#RUN go get -u github.com/golang/dep/cmd/dep
ADD ./main.go /go/src/app
#COPY Gopkg.toml /go/src/app

# Use distroless as minimal base image to package the vault-api-helper binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /go/src/app/app .
ENTRYPOINT ["/app"]
