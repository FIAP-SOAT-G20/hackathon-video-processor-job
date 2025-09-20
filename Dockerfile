# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/video-processor-job

# Final stage
FROM alpine:3.18

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Install FFmpeg
RUN apk update && \
    apk add --no-cache --virtual .build-deps wget xz tar && \
    wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz && \
    tar xf ffmpeg-release-amd64-static.tar.xz && \
    mv ffmpeg-*-amd64-static/ffmpeg /usr/local/bin/ && \
    mv ffmpeg-*-amd64-static/ffprobe /usr/local/bin/ && \
    rm -rf ffmpeg-* *.tar.xz && \
    apk del .build-deps && \
    rm -rf /var/cache/apk/*

WORKDIR /app
RUN chown -R appuser:appuser /app
USER appuser

COPY --from=build --chown=appuser:appuser /app/main ./main

ENTRYPOINT ["./main"]