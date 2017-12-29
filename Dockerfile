FROM golang
WORKDIR /go/src/github.com/samkreter/VSTSAutoReviewer/ 
COPY . /go/src/github.com/samkreter/VSTSAutoReviewer/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/samkreter/VSTSAutoReviewer/app .
CMD ["./app"]