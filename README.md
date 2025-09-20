# <p align="center"><b>Video Processor Job</b> <small>FIAP Hackathon - Video Frame Extraction Service</small></p>

<p align="center">
    <img src="https://img.shields.io/badge/Code-Go-informational?style=flat-square&logo=go&color=00ADD8" alt="Go" />
    <img src="https://img.shields.io/badge/Cloud-Kubernetes-informational?style=flat-square&logo=kubernetes&color=326CE5" alt="Kubernetes" />
    <img src="https://img.shields.io/badge/Storage-S3-informational?style=flat-square&logo=amazons3&color=569A31" alt="S3" />
    <img src="https://img.shields.io/badge/Tools-FFmpeg-informational?style=flat-square&logo=ffmpeg&color=007808" alt="FFmpeg" />
    <img src="https://img.shields.io/badge/Tools-Docker-informational?style=flat-square&logo=docker&color=2496ED" alt="Docker" />
    <img src="https://img.shields.io/badge/Tools-Make-informational?style=flat-square&logo=make&color=6D00CC" alt="Make" />
</p>

## 💬 About

Video processing job developed for the FIAP Hackathon that extracts frames from video files using FFmpeg. The service
follows Clean Architecture principles and runs as a Kubernetes job for scalable video processing.

### Key Features

- **Video Processing**: Extracts frames from uploaded video files at configurable frame rates
- **Frame Extraction**: Uses FFmpeg to extract frames as PNG or JPG images (configurable)
- **ZIP Compression**: Packages extracted frames into a downloadable ZIP file (uploaded to S3)
- **S3 Integration**: Downloads videos from S3, uploads processed frames, and cleans up original files
- **Kubernetes Job**: Runs as containerized Kubernetes job for scalable processing
- **Docker Support**: Containerized deployment with FFmpeg included
- **Clean Architecture**: Well-structured codebase following Clean Architecture principles

## 🏗️ Architecture

This service implements Clean Architecture with the following layers:

```
├── cmd/
│   └── video-processor-job/main.go # Kubernetes job entry point
├── internal/
│   ├── core/                   # Business logic layer
│   │   ├── domain/entity/      # Business entities
│   │   ├── dto/                # Data transfer objects
│   │   ├── port/               # Interface contracts
│   │   └── usecase/            # Business use cases
│   ├── adapter/                # Interface adapters
│   │   ├── controller/         # Request handlers
│   │   ├── gateway/            # External service interfaces
│   │   └── presenter/          # Response formatting
│   └── infrastructure/         # External concerns
│       ├── datasource/         # Data access implementations
│       ├── service/            # External service implementations
│       └── logger/             # Logging utilities
```

## 🔄 Processing Flow

1. **Input**: Kubernetes job receives environment variables with video S3 key
2. **Download**: Video file downloaded from S3 to container's `/tmp` directory
3. **Validation**: Video format validation using FFprobe
4. **Processing**: Frame extraction using FFmpeg at specified frame rate (default: 1 fps)
5. **Compression**: Extracted frames packaged into ZIP file
6. **Upload**: ZIP file uploaded to processed images S3 bucket
7. **Cleanup**: Original video deleted from source bucket and temporary files cleaned up
8. **Response**: Returns processing result with output location and frame count

## 🚀 Getting Started

### Prerequisites

- Go 1.25+
- FFmpeg (for local development)
- AWS CLI configured
- Docker (for containerized builds)
- Kubernetes cluster access

### Environment Variables

The application uses environment variables for configuration. See `.env.example` for a complete list.

**Core Variables:**

```bash
# AWS Credentials
K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key-id         # AWS Access Key
K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-access-key # AWS Secret Key
K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token         # AWS Session Token
K8S_JOB_ENV_AWS_REGION=us-east-1                         # AWS Region

# Video Processing
K8S_JOB_ENV_VIDEO_KEY=videos/sample.mp4                  # S3 key of video to process
K8S_JOB_ENV_VIDEO_BUCKET=your-raw-video-bucket          # S3 bucket for input videos
K8S_JOB_ENV_PROCESSED_BUCKET=your-processed-images-bucket # S3 bucket for output ZIP files
K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg                      # Output format (jpg or png)
K8S_JOB_ENV_VIDEO_EXPORT_FPS=1.0                         # Frame extraction rate
```

### Local Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Build video processor job binary
go build -o video-processor-job ./cmd/video-processor-job

# Run video processor job locally with environment variables
export K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key
export K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-key
export K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token
export K8S_JOB_ENV_AWS_REGION=us-east-1
export K8S_JOB_ENV_VIDEO_KEY=videos/sample.mp4
export K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg
export K8S_JOB_ENV_VIDEO_EXPORT_FPS=1.0
./video-processor-job
```

### Docker Build

```bash
# Build video processor job Docker image
docker build -t video-processor-job .

# Push to ECR for Kubernetes deployment
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPOSITORY
docker tag video-processor-job:latest $ECR_REPOSITORY:latest
docker push $ECR_REPOSITORY:latest
```

## Usage

### Video Processor Job

Run the video processor job for local development, testing, or Kubernetes deployment:

#### Command Line Usage

```bash
# Build video processor job binary
go build -o video-processor-job ./cmd/video-processor-job

# Set environment variables and run
export K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key
export K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-key
export K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token
export K8S_JOB_ENV_AWS_REGION=us-east-1
export K8S_JOB_ENV_VIDEO_KEY=videos/sample.mp4
export K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg
export K8S_JOB_ENV_VIDEO_EXPORT_FPS=1.0
export K8S_JOB_ENV_VIDEO_BUCKET=your-raw-video-bucket
export K8S_JOB_ENV_PROCESSED_BUCKET=your-processed-images-bucket
./video-processor-job
```

#### Docker Usage

```bash
# Build video processor job Docker image
docker build -t video-processor-job .

# Run with Docker (requires AWS credentials)
docker run --rm \
  -e K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key \
  -e K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token \
  -e K8S_JOB_ENV_AWS_REGION=us-east-1 \
  -e K8S_JOB_ENV_VIDEO_KEY=videos/sample.mp4 \
  -e K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg \
  -e K8S_JOB_ENV_VIDEO_EXPORT_FPS=2.0 \
  -e K8S_JOB_ENV_VIDEO_BUCKET=your-raw-video-bucket \
  -e K8S_JOB_ENV_PROCESSED_BUCKET=your-processed-images-bucket \
  video-processor-job
```

#### Video Processor Job Environment Variables

**Required:**

- `K8S_JOB_ENV_AWS_ACCESS_KEY_ID`: AWS Access Key ID
- `K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY`: AWS Secret Access Key
- `K8S_JOB_ENV_AWS_SESSION_TOKEN`: AWS Session Token (for temporary credentials)
- `K8S_JOB_ENV_VIDEO_KEY`: S3 key (path) of the video file to process
- `K8S_JOB_ENV_VIDEO_BUCKET`: S3 bucket name for input videos (default: video-processor-raw-videos)
- `K8S_JOB_ENV_PROCESSED_BUCKET`: S3 bucket name for output files (default: video-processor-processed-images)

**Optional:**
- `K8S_JOB_ENV_VIDEO_EXPORT_FORMAT`: Output format (`jpg` or `png`, default: `jpg`)
- `K8S_JOB_ENV_VIDEO_EXPORT_FPS`: Frame extraction rate (default: `1.0`)


### Response Format

The video processor job returns the following JSON response format:

```json
{
  "success": true,
  "message": "Video processed successfully. 3 frames extracted.",
  "output_key": "processed/0c1d1aa6f0fdd8cd33db975a87b77545a711a9a867ccb1270b9e583174f3c1b1.zip",
  "frame_count": 3,
  "hash": "0c1d1aa6f0fdd8cd33db975a87b77545a711a9a867ccb1270b9e583174f3c1b1"
}
```

**Note**: The `output_key` is the S3 object key for the generated ZIP in the processed bucket. The filename is based on
the SHA-256 hash of the original video content. Construct a public URL as needed (e.g., via a CDN or S3 access policy).

## 📝 Example Usage

### Video Processor Job with Environment Variables

**Option 1: Using .env file**

```bash
# Copy and configure environment file
cp .env.example .env
# Edit .env with your values

# Load environment and run
source .env
go build -o video-processor-job ./cmd/video-processor-job
./video-processor-job
```

**Option 2: Manual export**

```bash
# Set all required environment variables
export K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key
export K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-key
export K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token
export K8S_JOB_ENV_AWS_REGION=us-east-1
export K8S_JOB_ENV_VIDEO_KEY=videos/my-video.mp4
export K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg
export K8S_JOB_ENV_VIDEO_EXPORT_FPS=2.0
export K8S_JOB_ENV_VIDEO_BUCKET=my-raw-videos
export K8S_JOB_ENV_PROCESSED_BUCKET=my-processed-images

# Build and run
go build -o video-processor-job ./cmd/video-processor-job
./video-processor-job
```

### Docker with Environment Variables

**Option 1: Using .env file**

```bash
# Copy and configure environment file
cp .env.example .env
# Edit .env with your values

# Run with env file
docker run --rm --env-file .env video-processor-job
```

**Option 2: Manual environment variables**

```bash
docker run --rm \
  -e K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key \
  -e K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token \
  -e K8S_JOB_ENV_AWS_REGION=us-east-1 \
  -e K8S_JOB_ENV_VIDEO_KEY=videos/sample.mp4 \
  -e K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg \
  -e K8S_JOB_ENV_VIDEO_EXPORT_FPS=1.0 \
  -e K8S_JOB_ENV_VIDEO_BUCKET=video-processor-raw-videos \
  -e K8S_JOB_ENV_PROCESSED_BUCKET=video-processor-processed-images \
  video-processor-job
```

## 🛠️ Development

### Project Structure

- **Domain Layer**: Pure business logic without external dependencies
- **Application Layer**: Use cases orchestrating business workflows
- **Interface Layer**: Controllers, presenters, and gateways
- **Infrastructure Layer**: External services, databases, and frameworks

### Key Components

- **VideoProcessor**: FFmpeg integration for frame extraction
- **FileManager**: Temporary file management in container environment
- **StorageDataSource**: S3 operations for video and frame storage
- **VideoController**: Environment variable processing and response formatting

## 🧪 Testing

### Unit tests (with testify)

- Run unit tests: `make unit-test`
- Check coverage (fails if < 80%): `make coverage-check`

### BDD tests (Godog + testify)

- Define scenarios in `tests/features/*.feature`
- Run BDD tests: `make bdd-test`

### Mocks (Uber mockgen)

- Generate mocks for ports: `make mock`
- The generator uses `go.uber.org/mock/mockgen`. Generated files are stored under `internal/core/port/mocks`.

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests (requires AWS credentials)
make test-integration
```

## 📦 Deployment

### CI/CD (GitHub Actions)

This repo includes workflows to enforce quality and ship images:

- `ci-unit-test.yml`: runs unit tests and enforces >= 80% coverage
- `ci-bdd-tests.yml`: runs BDD tests with Godog
- `ci-build-deploy.yml`: builds and pushes the Docker image to ECR after BDD passes on `main`

Required repository secrets:

- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN` (if applicable)
- `AWS_REGION` (e.g., `us-east-1`)
- `ECR_REPOSITORY` (e.g., `fiap-hackathon-dev-video-processor`)

Images are tagged as `latest` and `${{ github.sha }}`.

### Docker Container

#### Video Processor Job Container

```bash
# Build video processor job image (Alpine Linux base)
docker build -t video-processor-job .

# Run video processor job container
docker run --rm \
  -e K8S_JOB_ENV_AWS_ACCESS_KEY_ID=your-access-key \
  -e K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e K8S_JOB_ENV_AWS_SESSION_TOKEN=your-session-token \
  -e K8S_JOB_ENV_AWS_REGION=us-east-1 \
  -e K8S_JOB_ENV_VIDEO_KEY=videos/test.mp4 \
  -e K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg \
  -e K8S_JOB_ENV_VIDEO_EXPORT_FPS=1.0 \
  -e K8S_JOB_ENV_VIDEO_BUCKET=your-bucket \
  -e K8S_JOB_ENV_PROCESSED_BUCKET=your-processed-bucket \
  video-processor-job
```

#### Push to ECR

```bash
# Push to ECR for Kubernetes deployment
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPOSITORY
docker tag video-processor-job:latest $ECR_REPOSITORY:latest
docker push $ECR_REPOSITORY:latest
```

## 🔧 Configuration

### Processing Configuration

- **frame_rate**: Frames per second to extract (default: 1.0)
- **output_format**: Output image format (default: "png", supports: png, jpg)

### S3 Bucket Structure

```
raw-videos/
├── video1.mp4
├── video2.avi
└── ...

processed-images/  
├── processed/video1_frames.zip
├── processed/video2_frames.zip
└── ...
```