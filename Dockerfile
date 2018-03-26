ARG GO_VERSION=1.10
FROM golang:${GO_VERSION}-alpine AS build-stage
WORKDIR /go/src/github.com/tony24681379/k8s-alert-controller
COPY ./ /go/src/github.com/tony24681379/k8s-alert-controller
RUN go test $(go list ./... | grep -v /vendor/) \
  && go install

FROM alpine:3.5
ENV TZ Asia/Taipei
COPY --from=build-stage /go/bin/k8s-alert-controller .
ENTRYPOINT ["/k8s-alert-controller"]