FROM golang:1.15-alpine AS builder
WORKDIR /src
COPY . .
RUN GO111MODULE=on go build

FROM alpine:latest AS server
WORKDIR /app
COPY --from=builder /src/isabelle ./isabelle

CMD [ "./isabelle" ]