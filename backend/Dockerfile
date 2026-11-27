FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY backend/go.mod backend/go.sum* ./backend/
WORKDIR /app/backend
RUN go mod download

COPY backend/ ./
COPY data/ ../data/

RUN CGO_ENABLED=0 go build -o /saycanto-api ./cmd/api && \
    CGO_ENABLED=0 go build -o /saycanto-importer ./cmd/importer && \
    /saycanto-importer -custom ../data/custom.json -out /dictionary.db

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /saycanto-api /app/saycanto-api
COPY --from=builder /dictionary.db /app/dictionary.db

ENV DB_PATH=/app/dictionary.db
ENV PORT=8080
EXPOSE 8080
CMD ["/app/saycanto-api"]
