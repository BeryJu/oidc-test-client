FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o oidc-test-client main.go

FROM debian:12-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /web-root

COPY --from=builder /app/oidc-test-client ./
COPY init.sh ./
RUN chmod +x init.sh oidc-test-client

# Expose the application port
EXPOSE 9009

# Set environment variables
ENV OIDC_BIND=0.0.0.0:9009

# Add a healthcheck to verify the application is running
HEALTHCHECK --interval=5s --start-period=1s CMD ["/web-root/oidc-test-client", "healthcheck"]

# Set the entrypoint to the init script
ENTRYPOINT ["sh", "init.sh"]
