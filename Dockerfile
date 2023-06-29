# syntax=docker/dockerfile:1

# Build the application from source
FROM docker.io/golang:1.19 AS build-stage

WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./
RUN go mod download -json

# Check the files and size
RUN ls -al
RUN df -h /app

COPY . .
RUN ls -al

RUN go build -a -o /k8smetrics_agent agent/*.go

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /k8smetrics_agent /k8smetrics_agent

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/k8smetrics_agent"]

