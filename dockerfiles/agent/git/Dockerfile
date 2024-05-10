FROM cgr.dev/chainguard/wolfi-base AS builder
RUN apk update && apk add go-1.21
WORKDIR /
COPY ./ ./

RUN go mod download
RUN CGO_ENABLED=0 go build -o ./build/agent agent/git/main.go

FROM cgr.dev/chainguard/static:latest
COPY --from=builder ./build/agent agent

USER 65532:65532
ENTRYPOINT ["/agent"]
