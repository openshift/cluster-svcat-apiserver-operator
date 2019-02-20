FROM registry.svc.ci.openshift.org/openshift/release:golang-1.10 AS builder
WORKDIR /go/src/github.com/openshift/cluster-svcat-apiserver-operator
COPY . .
RUN go build ./cmd/cluster-svcat-apiserver-operator

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base
COPY --from=builder /go/src/github.com/openshift/cluster-svcat-apiserver-operator/cluster-svcat-apiserver-operator /usr/bin/
COPY manifests /manifests
LABEL io.openshift.release.operator true
