FROM golang:1.9.2 as builder
WORKDIR  /go/src/github.com/samkreter/vstsautoreviewer/
COPY . /go/src/github.com/samkreter/vstsautoreviewer/
RUN go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service ./cmd/service

FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/samkreter/vstsautoreviewer/service .
ENTRYPOINT ["./service"]