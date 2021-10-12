FROM golang:latest AS builder
WORKDIR $GOPATH/src/beryju.org/oidc-test-client
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -v -o /go/bin/oidc-test-client

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/oidc-test-client /oidc-test-client
EXPOSE 9009
WORKDIR /web-root
ENV OIDC_BIND=0.0.0.0:9009
HEALTHCHECK CMD [ "wget", "--spider", "http://localhost:9009/health" ]
CMD [ "/oidc-test-client" ]
ENTRYPOINT [ "/oidc-test-client" ]
