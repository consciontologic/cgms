FROM alpine@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
COPY .local/bin/server /app/server
WORKDIR /app
USER 1000:1000
ENTRYPOINT ["/app/server"]
HEALTHCHECK --interval=5s --timeout=2s --retries=6 CMD wget -q -O /dev/null http://127.0.0.1:8080/health/ready || exit 1
