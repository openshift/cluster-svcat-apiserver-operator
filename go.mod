module github.com/openshift/cluster-svcat-apiserver-operator

go 1.13

require (
	github.com/openshift/api v0.0.0-20200217161739-c99157bc6492
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/sirupsen/logrus v1.4.2
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.3-beta.0
	k8s.io/client-go v0.17.2
)

replace golang.org/x/text => golang.org/x/text v0.3.3
