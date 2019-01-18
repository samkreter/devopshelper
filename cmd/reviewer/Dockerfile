FROM golang:1.9.2 as builder
WORKDIR  /go/src/github.com/samkreter/VSTSAutoReviewer/
COPY . /go/src/github.com/samkreter/VSTSAutoReviewer/
RUN go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reviewer ./cmd/reviewer

FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/samkreter/VSTSAutoReviewer/reviewer .
ENTRYPOINT ["./reviewer"]