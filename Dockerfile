FROM golang:1.19-alpine AS builder
WORKDIR /go/src/smartmeter
COPY . .
RUN go mod download
RUN go build -o app . && ls -al .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/smartmeter .
EXPOSE 8080
CMD ["./app"]  