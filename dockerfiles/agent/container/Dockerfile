FROM cgr.dev/chainguard/go:1.20 AS base
WORKDIR /
COPY ./ ./

RUN go mod download \
    && CGO_ENABLED=0 go build -o ./build/agent agent/container/main.go

FROM scratch
COPY --from=base ./build/agent agent

USER 65532:65532
ENTRYPOINT [ "/agent" ]
