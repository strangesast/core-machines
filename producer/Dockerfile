FROM golang:alpine as build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir /build && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /build .

FROM scratch
COPY --from=build /build /app
CMD ["/app/serial-monitoring"]
