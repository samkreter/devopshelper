FROM golang:1.9.2 as builder
WORKDIR  /go/src/images/vstsAutoReviewer/
COPY . /go/src/images/vstsAutoReviewer/
RUN go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reviewer ./cmd/reviewer

FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/images/vstsAutoReviewer/reviewer .
ENTRYPOINT ["./reviewer"]