FROM alpine:3.24.1

LABEL org.opencontainers.image.source="https://github.com/welovemedia/ffmate"
LABEL org.opencontainers.image.description="FFmate is a modern and powerful automation layer built on top of FFmpeg — designed to make video and audio transcoding simpler, smarter, and easier to integrate."
LABEL org.opencontainers.image.licenses="AGPL-3.0"

WORKDIR /app
RUN set -eux; \
    echo "Architecture:"; \
    apk --print-arch

RUN set -eux; \
    echo "Repositories:"; \
    cat /etc/apk/repositories

RUN set -eux; \
    echo "Updating indexes:"; \
    apk update

RUN set -eux; \
    echo "Installing bash:"; \
    apk add --no-cache bash

RUN set -eux; \
    echo "Installing jq:"; \
    apk add --no-cache jq

RUN set -eux; \
    echo "Installing curl:"; \
    apk add --no-cache curl

RUN set -eux; \
    echo "Installing jellyfin-ffmpeg:"; \
    apk add --no-cache jellyfin-ffmpeg

RUN set -eux; \
    echo "jellyfin-ffmpeg directory:"; \
    ls -la /usr/lib/jellyfin-ffmpeg

RUN set -eux; \
    echo "ffmpeg binary:"; \
    ls -la /usr/lib/jellyfin-ffmpeg/ffmpeg

RUN set -eux; \
    echo "ffprobe binary:"; \
    ls -la /usr/lib/jellyfin-ffmpeg/ffprobe

RUN set -eux; \
    echo "Linking ffmpeg:"; \
    ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/local/bin/ffmpeg

RUN set -eux; \
    echo "Linking ffprobe:"; \
    ln -s /usr/lib/jellyfin-ffmpeg/ffprobe /usr/local/bin/ffprobe

RUN set -eux; \
    echo "FFmpeg location:"; \
    command -v ffmpeg

RUN set -eux; \
    echo "FFmpeg version:"; \
    ffmpeg -version

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/ffmate /app/ffmate

RUN chmod 755 /app/ffmate

ENV PORT=3000 \
    DATABASE=/app/db/sqlite.db \
    DEBUGO="info:?,warn:?,error:?" \
    MAX_CONCURRENT_TASKS=3 \
    IDENTIFIER="" \
    GOCACHE=off

EXPOSE ${PORT}

RUN mkdir -p /app/db

CMD ["sh", "-c", "exec /app/ffmate server --port=\"$PORT\" --identifier=\"$IDENTIFIER\" --debug=\"$DEBUGO\" --database=\"$DATABASE\" --max-concurrent-tasks=\"$MAX_CONCURRENT_TASKS\""]