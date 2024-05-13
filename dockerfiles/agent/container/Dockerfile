FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /
COPY ./ ./

RUN go mod download
RUN CGO_ENABLED=0 go build -o ./build/agent agent/container/main.go

FROM cgr.dev/chainguard/static:latest
COPY --from=builder ./build/agent agent

USER 65532:65532
ENTRYPOINT ["/agent"]
