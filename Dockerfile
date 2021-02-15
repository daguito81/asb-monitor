FROM golang:1.15.6 as build

WORKDIR /sbmonitor
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o start main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/sbmonitor
COPY --from=build /sbmonitor .
CMD ["./start", "MAIN_SB"]