FROM docker.io/library/alpine:3.19.0
RUN apk add --no-cache ca-certificates
COPY oidc-test-client /
EXPOSE 9009
WORKDIR /web-root
ENV OIDC_BIND=0.0.0.0:9009
HEALTHCHECK --interval=5s --start-period=1s CMD [ "wget", "--spider", "http://localhost:9009/health" ]
ENTRYPOINT [ "/oidc-test-client" ]
