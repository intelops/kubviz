FROM golang:1.22 AS builder
WORKDIR /
COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o ./build/graphqlserver graphqlserver/server.go

FROM scratch
COPY --from=builder ./build/graphqlserver server

USER 65532:65532
ENTRYPOINT ["/server"]
