FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:3.20
RUN adduser -D -u 10001 api
WORKDIR /app
COPY --from=build /api /app/api
COPY public ./public
USER api
EXPOSE 8080
ENTRYPOINT ["/app/api"]
