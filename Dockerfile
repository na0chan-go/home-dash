FROM golang:1.23-bookworm AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/server /app/server
COPY config /app/config

ENV APP_ADDR=:8080
ENV DB_PATH=/data/app.db
ENV GARBAGE_SCHEDULE_PATH=config/garbage_schedule.json

EXPOSE 8080
ENTRYPOINT ["/app/server"]
