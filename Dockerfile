# SayCanto API image (build from the repo root): `docker build -t saycanto-api .`
#
# The full dictionary (CC-Canto + CC-CEDICT + frequencies + word types, ~89 MB) is built offline
# by cmd/importer and published as a GitHub Release asset, because its source files cannot all be
# downloaded by a script (MDBG forbids scripted CC-CEDICT downloads). The build fetches that file
# and checks its SHA-256. Pass --build-arg DICT_URL= (empty) to build only the curated words
# from data/custom.json instead, e.g. for a quick local test.

FROM golang:1.23-alpine AS builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY data/custom.json /app/data/custom.json
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /saycanto-api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /saycanto-importer ./cmd/importer

ARG DICT_URL=https://github.com/kaho-dev/saycanto/releases/download/data-2026-11-27/dictionary.db.gz
ARG DICT_SHA256=f5699a332b650f7d7693e8013c08ec74b1442aab3fbb68285cc9ef86cc142607
RUN set -eu; \
    if [ -n "$DICT_URL" ]; then \
      wget -q -O /tmp/dictionary.db.gz "$DICT_URL"; \
      echo "$DICT_SHA256  /tmp/dictionary.db.gz" | sha256sum -c -; \
      gunzip -c /tmp/dictionary.db.gz > /dictionary.db; \
      rm /tmp/dictionary.db.gz; \
    else \
      /saycanto-importer -custom /app/data/custom.json -out /dictionary.db; \
    fi

FROM alpine:3.20
RUN adduser -D -H saycanto && mkdir /app && chown saycanto /app
WORKDIR /app
COPY --from=builder /saycanto-api /app/saycanto-api
COPY --from=builder --chown=saycanto /dictionary.db /app/dictionary.db
USER saycanto

# Cloud Run sets PORT; CORS_ORIGIN should be the website's address (e.g. https://saycanto.vercel.app).
ENV DB_PATH=/app/dictionary.db \
    PORT=8080 \
    CORS_ORIGIN=*
EXPOSE 8080
CMD ["/app/saycanto-api"]
