FROM golang:1.15.6 AS builder
WORKDIR /go/src/smartmeter
RUN go get -d -u -v github.com/roaldnefs/go-dsmr
RUN go get -d -u -v github.com/tarm/serial
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/smartmeter .
EXPOSE 8080
CMD ["./app"]  