FROM golang:1.24.6-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gateway cmd/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=build /app/gateway /app/gateway

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/gateway"]
