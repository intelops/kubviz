FROM golang:1.20 as builder

WORKDIR /workspace
COPY ./ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o migration cmd/cli/main.go

RUN chmod +x /workspace/script/wait-for-clickhouse.sh

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/migration .
COPY --from=builder /workspace/sql /sql
COPY --from=builder /workspace/script /script

USER 65532:65532

ENTRYPOINT ["/migration"]