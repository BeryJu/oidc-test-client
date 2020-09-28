# OIDC-test-client

![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/beryju/oidc-test-client?style=flat-square)
![Docker pulls](https://img.shields.io/docker/pulls/beryju/oidc-test-client.svg?style=flat-square)

This is a small, golang-based OIDC Client, to be used in End-to-end or other testing. It uses the github.com/coreos/go-oidc Library for the actual OIDC Logic.

This tool can be used to test the traditional Authorization Code Flow. It also tests OIDC Token Introspection, if your provider supports it.

This tool is full configured using environment variables.

## URLs

- `http://localhost:9009/health`: Healthcheck URL, used by the docker healtcheck.
- `http://localhost:9009/auth/callback`: OAuth Callback URL
- `http://localhost:9009/`: Test URL, initiated OAuth Code flow

## Configuration

- `OIDC_BIND`: Which address and port to bind to. Defaults to `0.0.0.0:9009`.
- `OIDC_CLIENT_ID`: OAuth2 Client ID to use
- `OIDC_CLIENT_SECRET`: OAuth2 Client Secret to use
- `OIDC_PROVIDER`: Optional URL that metadata is fetched from. The metadata is fetched on the first request to `/`
- `OIDC_ROOT_URL`: URL under which you access this Client.

## Running

This service is intended to run in a docker container

```
docker pull beryju/oidc-test-client
docke run -d --rm \
    -p 9009:9009 \
    -e OIDC_CLIENT_ID=test-id \
    -e OIDC_CLIENT_SECRET=test-secret \
    -e OIDC_PROVIDER=http://id.beryju.org/... \
    beryju/oidc-test-client
```

Or if you want to use docker-compose, use this in your `docker-compose.yaml`.

```yaml
version: '3.5'

services:
  oidc-test-client:
    image: beryju/oidc-test-client
    ports:
      - 9009:9009
    environment:
      OIDC_CLIENT_ID: test-id
      OIDC_CLIENT_SECRET: test-secret
      OIDC_PROVIDER: https://some.issuer.tld/
```
