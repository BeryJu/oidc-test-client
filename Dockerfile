# https://github.com/GoogleContainerTools/distroless/blob/main/examples/go/Dockerfile
FROM golang:1.21 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11
LABEL org.opencontainers.image.source="https://github.com/beryju/oidc-test-client"

COPY --from=build /go/bin/app /

EXPOSE 9009
ENV OIDC_BIND=0.0.0.0:9009

HEALTHCHECK --interval=5s --start-period=1s CMD [ "/app", "healthcheck" ]
CMD ["/app"]
