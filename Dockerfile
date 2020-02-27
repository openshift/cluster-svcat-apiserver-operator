FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13 AS builder
WORKDIR /go/src/github.com/openshift/cluster-svcat-apiserver-operator
COPY . .
RUN make build --warn-undefined-variables

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base
COPY --from=builder /go/src/github.com/openshift/cluster-svcat-apiserver-operator/cluster-svcat-apiserver-remover /usr/bin/
COPY manifests /manifests
LABEL io.openshift.release.operator true
