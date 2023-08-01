FROM golang:1.17.13-alpine3.16 AS builder

WORKDIR /go/app
ADD . /go/app
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN apk add make
RUN make all

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /go/app/bin /app/

CMD ["./TN-Manager"]
