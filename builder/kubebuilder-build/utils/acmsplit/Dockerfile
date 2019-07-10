FROM golang:1.12.6 as builder
RUN mkdir /go/src/app
WORKDIR /go/src/app
ENV GOPATH=/go
ENV GO111MODULE=on
RUN go mod init
#RUN go get -u github.com/golang/dep/cmd/dep
ADD ./main.go /go/src/app
#COPY Gopkg.toml /go/src/app

#RUN dep ensure
RUN go test -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app

FROM alpine:3.9
RUN apk add --no-cache ca-certificates
CMD ["./app"]
COPY --from=builder /go/src/app/app .
