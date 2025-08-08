FROM golang:1.24.6-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o auth cmd/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=build /app/auth /app/auth

USER nonroot:nonroot

EXPOSE 50051

ENTRYPOINT ["/app/auth"]
