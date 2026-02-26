FROM golang:1.26-alpine AS base
WORKDIR /app

FROM base AS dev
RUN go install github.com/air-verse/air@latest
CMD ["air"]

FROM base AS build
COPY . .
RUN go build -mod=vendor -o main cmd/api/main.go

FROM alpine:3.22.0 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main"]


