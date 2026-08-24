### Stage 1: Build Vue.js frontend
FROM node:24-alpine AS frontend
WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

### Stage 2: Build Go backend
# Pass --build-arg BUILD_TAGS=mysql to enable MySQL support.
FROM golang:1.27-alpine AS backend
ARG BUILD_TAGS=""
# Build version, surfaced in the dashboard footer. Pass --build-arg VERSION=v1.2.3.
ARG VERSION=dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/web/dist ./internal/dashboard/dist
RUN CGO_ENABLED=0 go build -tags "${BUILD_TAGS}" -ldflags="-s -w -X main.version=${VERSION}" -o /gatecha ./cmd/gatecha

### Stage 3: Final runtime image
FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -h /app gatecha
WORKDIR /app
COPY --from=backend /gatecha .
RUN mkdir -p /app/data && chown gatecha:gatecha /app/data
USER gatecha
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["./gatecha"]
