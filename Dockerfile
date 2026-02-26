# Stage 1: Build Hugo site
FROM klakegg/hugo:0.111.3-ext-alpine AS hugo-builder

WORKDIR /src
COPY . .
RUN hugo --minify

# Stage 2: Build Go server
FROM golang:1.26-alpine AS go-builder

WORKDIR /app
COPY server/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Stage 3: Runtime
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy built artifacts
COPY --from=hugo-builder /src/public ./public
COPY --from=go-builder /app/server .

# Cloud Run expects port 8080
ENV PORT=8080
EXPOSE 8080

# Run the server
CMD ["./server"]
