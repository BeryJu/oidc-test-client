# OIDC-test-client

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/beryju/oidc-test-client/ci-build?style=for-the-badge)

This is a small, golang-based OIDC Client, to be used in End-to-end or other testing. It uses the github.com/coreos/go-oidc Library for the actual OIDC Logic.

This tool can be used to test the traditional Authorization Code Flow. It also tests OIDC Token Introspection, if your provider supports it.

This tool is full configured using environment variables.

## URLs

- `http://localhost:9009/health`: Healthcheck URL, used by the docker healtcheck.
- `http://localhost:9009/auth/callback`: OAuth Callback URL
- `http://localhost:9009/`: Test URL, initiated OAuth Code flow
- `http://localhost:9009/implicit/`: Tests an Implicit OIDC flow using `id_token token`

## Configuration

- `OIDC_BIND`: Which address and port to bind to. (defaults `0.0.0.0:9009`).
- `OIDC_CLIENT_ID`: OAuth2 Client ID to use.
- `OIDC_CLIENT_SECRET`: OAuth2 Client Secret to use. Can be set to an empty string when only implicit flow is tested.
- `OIDC_ROOT_URL`: URL under which you access this Client. (default http://localhost:9009)
- `OIDC_PROVIDER`: Optional URL that metadata is fetched from. The metadata is fetched on the first request to `/`
- `OIDC_SCOPES`: Scopes to request from the provider. Defaults to "openid,offline_access,profile,email"
- `OIDC_DO_REFRESH`: Whether refresh-token related checks are enabled (don't ask for a refresh token) (default: true)
- `OIDC_DO_INTROSPECTION`: Whether introspection related checks are enabled (don't call introspection endpoint) (default: true)
- `OIDC_DO_USER_INFO`: Whether user-info related checks are enabled (don't use userinfo endpoint) (default: true)
- `OIDC_TLS_VERIFY`: Whether to verify TLS certicates (set to "false" for self-signed) (default: true)

## Running

This service is intended to run in a docker container

```
# beryju.org is a vanity URL for ghcr.io/beryju
docker pull ghcr.io/beryju/oidc-test-client:
docke run -d --rm \
    -p 9009:9009 \
    -e OIDC_CLIENT_ID=test-id \
    -e OIDC_CLIENT_SECRET=test-secret \
    -e OIDC_PROVIDER=http://id.beryju.io/... \
    ghcr.io/beryju/oidc-test-client:
```

Or if you want to use docker-compose, use this in your `docker-compose.yaml`.

```yaml
version: '3.5'

services:
  oidc-test-client:
    image: ghcr.io/beryju/oidc-test-client:
    ports:
      - 9009:9009
    environment:
      OIDC_CLIENT_ID: test-id
      OIDC_CLIENT_SECRET: test-secret
      OIDC_PROVIDER: https://some.issuer.tld/
```
