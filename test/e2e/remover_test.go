package e2e

import (
	"testing"

	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	test "github.com/openshift/cluster-svcat-apiserver-operator/test/library"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

var removerNamespaceName = "openshift-service-catalog-apiserver-remover"
var operatorNamespaceName = "openshift-service-catalog-apiserver-operator"

func TestRemoverNamespace(t *testing.T) {
	kubeConfig, err := test.NewClientConfigForTest()
	if err != nil {
		t.Fatal(err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatal(err)
	}

	_, err = kubeClient.CoreV1().Namespaces().Get(removerNamespaceName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOperatorNamespaceRemoval(t *testing.T) {
	kubeConfig, err := test.NewClientConfigForTest()
	if err != nil {
		t.Fatal(err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatal(err)
	}

	_, err = kubeClient.CoreV1().Namespaces().Get(operatorNamespaceName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatalf("%s namespace was not removed", operatorNamespaceName)
	}
}

func TestOperatorCRRemoval(t *testing.T) {
	kubeConfig, err := test.NewClientConfigForTest()
	if err != nil {
		t.Fatal(err)
	}

	operatorClient, err := operatorclient.NewForConfig(kubeConfig)
	if err != nil {
		log.Errorf("problem getting operator client, error %v", err)
	}

	operatorConfigClient := operatorClient.OperatorV1()
	_, err = operatorConfigClient.ServiceCatalogAPIServers().Get("cluster", metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("ServiceCatalogAPIServer CR was not removed")
	}

}
