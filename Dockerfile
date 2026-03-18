# Stage 1: Build widget
FROM node:20-alpine AS widget-builder
WORKDIR /widget
COPY widget/package.json ./
RUN npm install
COPY widget/ ./
RUN npm run build

# Stage 2: Build Go server
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o blueprint-chat ./cmd/server

# Stage 3: Runtime (minimal image)
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/blueprint-chat .

# Widget dist
COPY --from=widget-builder /widget/dist ./widget/dist

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./blueprint-chat"]
