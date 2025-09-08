# <p align="center"><b>Video Processor Job</b> <small>FIAP Hackathon - Video Frame Extraction Service</small></p>

<p align="center">
    <img src="https://img.shields.io/badge/Code-Go-informational?style=flat-square&logo=go&color=00ADD8" alt="Go" />
    <img src="https://img.shields.io/badge/Cloud-AWS_Lambda-informational?style=flat-square&logo=awslambda&color=FF9900" alt="AWS Lambda" />
    <img src="https://img.shields.io/badge/Storage-S3-informational?style=flat-square&logo=amazons3&color=569A31" alt="S3" />
    <img src="https://img.shields.io/badge/Tools-FFmpeg-informational?style=flat-square&logo=ffmpeg&color=007808" alt="FFmpeg" />
    <img src="https://img.shields.io/badge/Tools-Docker-informational?style=flat-square&logo=docker&color=2496ED" alt="Docker" />
    <img src="https://img.shields.io/badge/Tools-Make-informational?style=flat-square&logo=make&color=6D00CC" alt="Make" />
</p>

## 💬 About

AWS Lambda function developed for the FIAP Hackathon that processes video files and extracts frames using FFmpeg. The service follows Clean Architecture principles and handles video processing in a serverless environment.

### Key Features

- **Video Processing**: Extracts frames from uploaded video files at configurable frame rates
- **Frame Extraction**: Uses FFmpeg to extract frames as PNG or JPG images (configurable)
- **ZIP Compression**: Packages extracted frames into a downloadable ZIP file (uploaded to S3)
- **S3 Integration**: Downloads videos from S3, uploads processed frames, and cleans up original files
- **Serverless**: Runs as AWS Lambda function with optimized performance
- **Clean Architecture**: Well-structured codebase following Clean Architecture principles

## 🏗️ Architecture

This service implements Clean Architecture with the following layers:

```
├── main.go                     # Entry point and dependency injection
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
│       ├── aws/lambda/         # Lambda-specific handlers
│       └── logger/             # Logging utilities
```

## 🔄 Processing Flow

1. **Input**: Lambda receives event with video S3 key
2. **Download**: Video file downloaded from S3 to Lambda's `/tmp` directory
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

### Environment Variables

```bash
VIDEO_BUCKET=your-raw-video-bucket          # S3 bucket for input videos
PROCESSED_BUCKET=your-processed-images-bucket # S3 bucket for output ZIP files
```

### Local Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Build binary
make build

# Run locally (requires video file)
make run
```

### Lambda Deployment

```bash
# Build Lambda deployment package
make build-lambda

# Deploy to AWS (requires terraform or SAM)
make deploy
```

## 📋 Usage

### Lambda Event Format

```json
{
  "video_key": "path/to/video.mp4",
  "configuration": {
    "frame_rate": 1.0,
    "output_format": "png"
  }
}
```

### Response Format

Note: output_key is the S3 object key for the generated ZIP in the processed bucket. Construct a public URL as needed (e.g., via a CDN or S3 access policy).

```json
{
  "success": true,
  "message": "Video processed successfully. 120 frames extracted.",
  "output_key": "processed/path/to/video_frames.zip",
  "frame_count": 120
}
```

## 🛠️ Development

### Project Structure

- **Domain Layer**: Pure business logic without external dependencies
- **Application Layer**: Use cases orchestrating business workflows  
- **Interface Layer**: Controllers, presenters, and gateways
- **Infrastructure Layer**: External services, databases, and frameworks

### Key Components

- **VideoProcessor**: FFmpeg integration for frame extraction
- **FileManager**: Temporary file management in Lambda environment
- **StorageDataSource**: S3 operations for video and frame storage
- **VideoController**: Lambda event handling and response formatting

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

```bash
# Build container image
make docker-build

# Push to ECR
make docker-push
```

### AWS Lambda

The service is designed to run as AWS Lambda function with:

- **Runtime**: Custom container or Go 1.x runtime
- **Memory**: 1024MB+ (for video processing)
- **Timeout**: 5-15 minutes (depending on video size)
- **Storage**: 512MB+ `/tmp` space for temporary files

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

## 📈 Performance

- **Cold Start**: ~2-3 seconds
- **Processing**: ~30-60 seconds per minute of video
- **Memory Usage**: Scales with video resolution and length
- **Concurrent Executions**: Configurable based on workload

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Team

FIAP SOAT G20 - Hackathon Team

---

<p align="center">Made with ❤️ for FIAP Hackathon</p>