FROM golang:1.20.0-alpine3.16 as build

COPY . /app
WORKDIR /app

RUN if [ ! -f go.mod ]; then go mod http-multiplexer; fi
RUN if [ ! -f go.sum ]; then go mod tidy; fi
RUN go build -o index

FROM alpine:3.16 as product
COPY --from=build /app/index .
CMD ["./index"]