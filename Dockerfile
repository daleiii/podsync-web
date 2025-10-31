FROM golang:1.25 AS builder

ENV TAG="nightly"
ENV COMMIT=""

WORKDIR /build

# Install Node.js for frontend build
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

COPY . .

RUN make build

# Download yt-dlp
RUN wget -O /usr/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && \
    chmod a+rwx /usr/bin/yt-dlp

# Alpine 3.22 will go EOL on 2027-05-01
FROM alpine:3.22

WORKDIR /app

# deno is required for yt-dlp (ref: https://github.com/yt-dlp/yt-dlp/issues/14404)
RUN apk --no-cache add ca-certificates python3 py3-pip ffmpeg tzdata libc6-compat deno wget

RUN chmod 777 /usr/local/bin
COPY --from=builder /usr/bin/yt-dlp /usr/local/bin/youtube-dl
COPY --from=builder /build/bin/podsync /app/podsync
COPY --from=builder /build/html /app/html

ENTRYPOINT ["/app/podsync"]
CMD ["--no-banner"]
