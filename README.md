# <p align="center"><b>Video Processor Job</b> <small>FIAP Hackathon — Video Frame Extraction Service</small></p>

<p align="center">
    <img src="https://img.shields.io/badge/Code-Go-informational?style=flat-square&logo=go&color=00ADD8" alt="Go" />
    <img src="https://img.shields.io/badge/Cloud-Kubernetes-informational?style=flat-square&logo=kubernetes&color=326CE5" alt="Kubernetes" />
    <img src="https://img.shields.io/badge/Storage-S3-informational?style=flat-square&logo=amazons3&color=569A31" alt="S3" />
    <img src="https://img.shields.io/badge/Tools-FFmpeg-informational?style=flat-square&logo=ffmpeg&color=007808" alt="FFmpeg" />
    <img src="https://img.shields.io/badge/Tools-Docker-informational?style=flat-square&logo=docker&color=2496ED" alt="Docker" />
    <img src="https://img.shields.io/badge/Tools-Make-informational?style=flat-square&logo=make&color=6D00CC" alt="Make" />
</p>

## 💬 Overview

A video processing job for the FIAP Hackathon. It downloads a video from S3, validates it with FFprobe, extracts frames
with FFmpeg, packages them into a ZIP, and uploads the result back to S3. The code follows Clean Architecture and can
run locally (binary/Docker) or as a Kubernetes Job.

### Key features

- Frame extraction (JPG/PNG) with configurable FPS
- ZIP packaging of extracted frames
- S3 integration (download, upload, cleanup)
- Container image with FFmpeg bundled
- Clean Architecture (well-defined layers)

## 🏗️ Architecture

```
├── cmd/
│   └── video-processor-job/main.go        # job entrypoint
├── internal/
│   ├── core/                              # business rules
│   │   ├── domain/entity/                 # entities
│   │   ├── dto/                           # DTOs
│   │   ├── port/                          # contracts (interfaces)
│   │   └── usecase/                       # use cases
│   ├── adapter/                           # interface adapters
│   │   ├── controller/                    # orchestration/input
│   │   ├── gateway/                       # external integrations
│   │   └── presenter/                     # response formatting
│   └── infrastructure/                    # infrastructure
│       ├── datasource/                    # S3, etc.
│       ├── service/                       # FFmpeg, files, etc.
│       └── logger/                        # logging
```

## 🔄 Processing flow

1. Receive the S3 video key via environment variables
2. Download the file to the container (e.g., /tmp)
3. Validate the video with FFprobe
4. Extract frames with FFmpeg at the configured FPS (default 1.0)
5. Zip extracted frames
6. Upload the ZIP to the processed bucket
7. Delete the original video from the source bucket and cleanup temporary files
8. Return a JSON result with success, frame count and output key

## ⚙️ Requirements
- Go 1.25+
- FFmpeg installed (only if running locally outside Docker)
- Docker (optional, recommended)
- AWS CLI configured (required for ECR and integration tests)

## 🔐 Environment variables

Refer to `.env.example` for the full list. Main variables:

Required:

- VIDEO_KEY
- VIDEO_BUCKET
- PROCESSED_BUCKET
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_SESSION_TOKEN

Optional (defaults):

- K8S_JOB_ENV_AWS_REGION (default: `us-east-1`)
- K8S_JOB_ENV_VIDEO_EXPORT_FORMAT (`jpg` or `png`, default: `jpg`)
- K8S_JOB_ENV_VIDEO_EXPORT_FPS (default: `1.0`)

Tip: use a `.env` file to avoid exposing secrets in commands (see below).

## 🚀 Quickstart

Local (binary):
```bash
make deps
make build
# configure your environment (edit .env first)
cp .env.example .env
source .env
./video-processor-job
```

Docker:
```bash
docker build -t video-processor-job .
# using an env file (recommended)
docker run --rm --env-file .env video-processor-job
```

Minimal example (without env-file; replace values):
```bash
docker run --rm \
  -e AWS_ACCESS_KEY_ID=... \
  -e AWS_SECRET_ACCESS_KEY=... \
  -e AWS_SESSION_TOKEN=... \
  -e AWS_REGION=us-east-1 \
  -e VIDEO_KEY=videos/sample.mp4 \
  -e VIDEO_EXPORT_FORMAT=jpg \
  -e VIDEO_EXPORT_FPS=1.0 \
  -e VIDEO_BUCKET=video-processor-raw-videos \
  -e PROCESSED_BUCKET=video-processor-processed-images \
  video-processor-job
```

## 📤 Response format

```json
{
  "success": true,
  "message": "Video processed successfully. 3 frames extracted.",
  "output_key": "processed/<hash>.zip",
  "frame_count": 3,
  "hash": "<video-sha256>"
}
```

The output filename is derived from the SHA-256 hash of the original video content.

## 🧪 Testing

- Lint: `make lint` (or `make lint-ci` for golangci-lint)
- Unit tests: `make unit-test`
- Coverage (fails if < 80%): `make coverage-check`
- Coverage report (HTML): `make test-coverage`
- Integration tests (require AWS credentials): `make test-integration`
- Generate mocks (uber mockgen): `make mock`
- Security scan: `make security-scan`

## 📦 CI/CD

Available GitHub Actions workflows:

- `.github/workflows/ci-unit-test.yml` — unit tests
- `.github/workflows/ci-go-test-coverage.yaml` — coverage report and PR comment
- `.github/workflows/ci-golangci-lint.yml` — linter
- `.github/workflows/ci-govulncheck.yml` — vulnerability check
- `.github/workflows/ci-integration-test.yml` — integration tests
- `.github/workflows/ci-build-deploy.yml` — build and push image to ECR (tags: `latest` and `${{ github.sha }}`)

Required repository secrets (for CI deploy):

- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
- `AWS_REGION` (e.g., `us-east-1`)
- `ECR_REPOSITORY` (e.g., `fiap-hackathon-dev-video-processor`)

Manual push to ECR (optional):
```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPOSITORY
docker tag video-processor-job:latest $ECR_REPOSITORY:latest
docker push $ECR_REPOSITORY:latest
```

## 🐳 Docker image

- Multi-stage build (Go 1.25)
- Alpine final image with static FFmpeg/FFprobe from johnvansickle.com
- Non-root user

## 🔧 Configuration notes

- frame_rate: extraction FPS (default 1.0)
- output_format: image format (default "jpg"; supports: `jpg`, `png`)

S3 bucket structure (defaults):
```
video-processor-raw-videos/
├── video1.mp4
├── video2.avi
└── ...

video-processor-processed-images/
├── processed/video1_frames.zip
├── processed/video2_frames.zip
└── ...
```
