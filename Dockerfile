# ======= STAGE 1: build =======
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/shortener-api
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/shortener-worker

# ======= STAGE 2 =======
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/worker /app/worker
EXPOSE 8003
CMD ["/app/api"]

