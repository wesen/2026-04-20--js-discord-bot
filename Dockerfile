FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --system --uid 10001 --home-dir /app --shell /usr/sbin/nologin appuser

WORKDIR /app

COPY .bin/discord-bot-linux-amd64 /usr/local/bin/discord-bot
COPY examples/discord-bots /app/examples/discord-bots

RUN chmod +x /usr/local/bin/discord-bot \
  && chown -R appuser:appuser /app

USER appuser
ENTRYPOINT ["discord-bot"]
