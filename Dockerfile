### Stage 1: Build Vue.js frontend
FROM node:20-alpine AS frontend
WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

### Stage 2: Build Go backend
FROM golang:1.26-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/web/dist ./internal/dashboard/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /gatecha ./cmd/gatecha

### Stage 3: Final runtime image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -h /app gatecha
WORKDIR /app
COPY --from=backend /gatecha .
USER gatecha
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["./gatecha"]
