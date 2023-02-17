FROM docker.io/library/golang:1.19
MAINTAINER Mario Vazquez <mavazque@redhat.com>
ARG gitCommit=notSet

WORKDIR /go/src/github.com/rhsyseng/cluster0-operators/
ADD cmd /go/src/github.com/rhsyseng/cluster0-operators/cmd
ADD pkg /go/src/github.com/rhsyseng/cluster0-operators/pkg
COPY go.mod /go/src/github.com/rhsyseng/cluster0-operators/
COPY go.sum /go/src/github.com/rhsyseng/cluster0-operators/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'github.com/rhsyseng/cluster0-operators/pkg/version.gitCommit=${gitCommit}' -X 'github.com/rhsyseng/cluster0-operators/pkg/version.buildTime=$(date +%Y-%m-%dT%H:%M:%SZ)'" -o ./out/cluster0-operators cmd/main.go

FROM scratch
MAINTAINER Mario Vazquez <mavazque@redhat.com>
COPY --from=0 /go/src/github.com/rhsyseng/cluster0-operators/out/cluster0-operators .
ENTRYPOINT ["/cluster0-operators"]
