FROM debian:bookworm-slim

WORKDIR /app

COPY mooltah* minijinja-cli* entrypoint.sh ./

ENTRYPOINT ["./entrypoint.sh"]