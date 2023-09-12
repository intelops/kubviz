# Build the manager binary
FROM golang:1.20 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY ./ ./
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o migration cmd/cli/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/migration .
COPY /workspace/sql /sql
COPY /workspace/script /script
RUN chmod +x /script/wait-for-clickhouse.sh
USER 65532:65532

ENTRYPOINT ["/migration"]
