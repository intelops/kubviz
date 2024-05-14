FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /
COPY ./ ./

RUN go mod download
RUN CGO_ENABLED=0 go build -o ./build/agent agent/git/main.go

FROM scratch
COPY --from=builder ./build/agent agent

USER 65532:65532
ENTRYPOINT ["/agent"]
