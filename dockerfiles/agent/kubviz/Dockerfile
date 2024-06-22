# Build the manager binary
FROM golang:1.22 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY ./ ./
RUN go mod download
RUN gofmt -w -r '"github.com/googleapis/gnostic/OpenAPIv2" -> "github.com/googleapis/gnostic/openapiv2"' /go/pkg/mod/sigs.k8s.io/kustomize/
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o kubviz_agent agent/kubviz/*.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM golang:alpine
WORKDIR /
COPY --from=builder /workspace/kubviz_agent .
COPY --from=zegl/kube-score:v1.16.0 /usr/bin/kube-score /usr/bin/kube-score
COPY --from=bitnami/kubectl:1.22.5 /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/
COPY --from=busybox:1.35.0-uclibc /bin/sh /bin/sh
COPY --from=aquasec/trivy:v0.43.1 /usr/local/bin/trivy /usr/local/bin/trivy

USER 65532:65532

ENTRYPOINT ["/kubviz_agent"]
