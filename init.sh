#!/bin/bash
set -e

case "$1" in
  "web")
    exec ./oidc-test-client
    ;;
esac
