FROM node:22-bookworm AS web-builder
WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM golang:1.23-bookworm AS go-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/web/dist /app/web/dist

RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /out/server /app/server
COPY --from=go-builder /app/config /app/config
COPY --from=go-builder /app/web/dist /app/web/dist

ENV APP_ADDR=:8080
ENV DB_PATH=/data/app.db
ENV GARBAGE_SCHEDULE_PATH=config/garbage_schedule.json
ENV WEB_DIST_PATH=web/dist
ENV CORS_ALLOW_ORIGINS=

EXPOSE 8080
ENTRYPOINT ["/app/server"]
