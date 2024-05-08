FROM cgr.dev/chainguard/wolfi-base AS builder
RUN apk update && apk add go-1.21
WORKDIR /
COPY ./ ./

RUN CGO_ENABLED=0 go build -o ./build/agent agent/container/main.go

FROM cgr.dev/chainguard/static:latest
COPY --from=builder ./build/agent agent

USER 65532:65532
ENTRYPOINT ["/agent"]
