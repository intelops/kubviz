FROM docker.io/golang:1.19 AS build-stage

WORKDIR /workspace
# Copy the Go Modules manifests
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -json
COPY . .
RUN ls -al
# Build
RUN go build -o /k8smetrics_agent agent/*.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build-stage /k8smetrics_agent /k8smetrics_agent

ENTRYPOINT ["/k8smetrics_agent"]
