FROM golang:alpine AS builder

RUN apk update && apk add ca-certificates && apk add protobuf protobuf-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir proto && go get -u github.com/golang/protobuf/protoc-gen-go && \
  protoc --go_out=plugins=grpc:proto ./monitoring.proto
RUN go build -o ./out/main .

FROM alpine
COPY --from=builder /app/out/main /app/main
EXPOSE 8080
CMD ["/app/main"]
