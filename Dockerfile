FROM golang:1.12-alpine3.9 AS builder
RUN apk add git
RUN go get -v github.com/gorilla/handlers
RUN go get -v github.com/gorilla/mux
COPY ./main.go ./main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build  -v -a -installsuffix cgo -o main .
FROM balenalib/rpi-alpine AS main
COPY --from=builder /go/main ./main
COPY ./static ./static
CMD ["./main"]
