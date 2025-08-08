FROM golang:1.24.6-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o notification cmd/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=build /app/notification /app/notification

USER nonroot:nonroot

ENTRYPOINT ["/app/notification"]
