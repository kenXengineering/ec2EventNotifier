FROM golang:alpine3.8
WORKDIR /go/src/github.com/kenXengineering/ec2EventNotifier
COPY . .
RUN CGO_ENABLE=0 GOOS=linux go build -o ec2EventNotifier -v .

FROM alpine:3.8
COPY --from=0 /go/src/github.com/kenXengineering/ec2EventNotifier /bin/ec2EventNotifier
RUN apk add --no-cache ca-certificates
ENTRYPOINT ["ec2EventNotifier"]