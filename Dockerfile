FROM alpine:3.24.1

LABEL org.opencontainers.image.source="https://github.com/welovemedia/ffmate"
LABEL org.opencontainers.image.description="FFmate is a modern and powerful automation layer built on top of FFmpeg — designed to make video and audio transcoding simpler, smarter, and easier to integrate."
LABEL org.opencontainers.image.licenses="AGPL-3.0"

WORKDIR /app

RUN apk add --no-cache jellyfin-ffmpeg bash jq curl

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/ffmate /app/ffmate

RUN ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/local/bin/ffmpeg \
 && ln -s /usr/lib/jellyfin-ffmpeg/ffprobe /usr/local/bin/ffprobe
 
ENV PORT=3000 \
    DATABASE=/app/db/sqlite.db \
    DEBUGO="info:?,warn:?,error:?" \
    MAX_CONCURRENT_TASKS=3 \
    IDENTIFIER=

EXPOSE ${PORT}

RUN mkdir -p /app/db

CMD ["sh", "-c", "/app/ffmate server --port=\"$PORT\" --identifier=\"$IDENTIFIER\" --debug=\"$DEBUGO\" --database=\"$DATABASE\" --max-concurrent-tasks=\"$MAX_CONCURRENT_TASKS\""]