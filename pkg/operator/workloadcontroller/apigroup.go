package workloadcontroller

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/rest"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var apiServiceGroupVersions = []schema.GroupVersion{
	// these are all the apigroups we manage
	{Group: "servicecatalog.k8s.io", Version: "v1beta1"},
}

func checkForAPIs(restclient rest.Interface, groupVersions ...schema.GroupVersion) []string {
	missingMessages := []string{}
	for _, groupVersion := range groupVersions {
		url := "/apis/" + groupVersion.Group + "/" + groupVersion.Version

		statusCode := 0
		restclient.Get().AbsPath(url).Do().StatusCode(&statusCode)
		if statusCode != http.StatusOK {
			missingMessages = append(missingMessages, fmt.Sprintf("%s.%s is not ready: %v", groupVersion.Version, groupVersion.Group, statusCode))
		}
	}

	return missingMessages
}

func APIServiceReferences() []configv1.ObjectReference {
	ret := []configv1.ObjectReference{}
	for _, gv := range apiServiceGroupVersions {
		ret = append(ret, configv1.ObjectReference{Group: "apiregistration.k8s.io", Resource: "apiservices", Name: gv.Version + "." + gv.Group})
	}
	return ret
}
