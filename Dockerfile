# Build stage
FROM golang:1.21 AS build

WORKDIR /app

# Copy only go.mod and go.sum to leverage Docker layer caching
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Ensure go.sum is up to date
RUN go mod tidy

# Build the Go binary for AWS Lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -ldflags="-w -s" -o main .

# Final image based on AWS Lambda custom runtime base
FROM public.ecr.aws/lambda/provided:al2023

# Install FFmpeg for video processing
RUN dnf update -y && \
    dnf install -y ffmpeg && \
    dnf clean all

# Copy the compiled binary from the build stage
COPY --from=build /app/main ./main

# Set the Lambda entry point
ENTRYPOINT ["./main"]