FROM golang:1.17.13 AS builder

WORKDIR /go/app
ADD . /go/app
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN apt install make -y
RUN make all

FROM ubuntu:20.04
WORKDIR /app
COPY --from=builder /go/app/bin /app/
RUN apt update && apt install bridge-utils iproute2 -y

CMD ["./TN-Manager"]
