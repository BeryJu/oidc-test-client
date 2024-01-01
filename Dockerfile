FROM docker.io/library/debian:12-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean
COPY oidc-test-client /
EXPOSE 9009
WORKDIR /web-root
ENV OIDC_BIND=0.0.0.0:9009
HEALTHCHECK --interval=5s --start-period=1s CMD [ "/oidc-test-client", "healthcheck" ]
ENTRYPOINT [ "/oidc-test-client" ]
