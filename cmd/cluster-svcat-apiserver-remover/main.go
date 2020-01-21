package main

import (
	operatorapiv1 "github.com/openshift/api/operator/v1"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var targetNamespaceName = "openshift-service-catalog-apiserver-operator"

func createClientConfigFromFile(configPath string) (*rest.Config, error) {
	clientConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	log.Info("Starting openshift-service-catalog-apiserver-remover job")

	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		clientConfig, err = createClientConfigFromFile(homedir.HomeDir() + "/.kube/config")
		if err != nil {
			log.Error("Failed to create LocalClientSet")
			panic(err.Error())
		}
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err.Error())
	}

	operatorClient, err := operatorclient.NewForConfig(clientConfig)
	if err != nil {
		log.Errorf("problem getting operator client, error %v", err)
	}
	operatorConfigClient := operatorClient.OperatorV1()
	operatorConfig, err := operatorConfigClient.ServiceCatalogAPIServers().Get("cluster", metav1.GetOptions{})
	if err != nil {
		log.Errorf("problem getting ServiceCatalogAPIServer CR, error %v", err)
	}

	switch operatorConfig.Spec.ManagementState {
	case operatorapiv1.Managed:
		log.Warning("We found a cluster-svcat-apiserver-operator in Managed state. Aborting")
	case operatorapiv1.Unmanaged:
		log.Info("ServiceCatalogAPIServer managementState is 'Unmanaged'")
		log.Infof("Removing target namespace %s", targetNamespaceName)
		if err := kubeClient.CoreV1().Namespaces().Delete(targetNamespaceName, nil); err != nil && !apierrors.IsNotFound(err) {
			log.Errorf("problem removing target namespace [%s] :  %v", targetNamespaceName, err)
		}
		log.Info("Removing the ServiceCatalogAPIServer CR")
		err = operatorConfigClient.ServiceCatalogAPIServers().Delete("cluster", &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("ServiceCatalogAPIServer cr deletion failed: %v", err)
		} else {
			log.Info("ServiceCatalogAPIServer cr removed successfully.")
		}
	case operatorapiv1.Removed:
		log.Info("ServiceCatalogAPIServer managementState is 'Removed'")
		log.Infof("Removing target namespace %s", targetNamespaceName)
		if err := kubeClient.CoreV1().Namespaces().Delete(targetNamespaceName, nil); err != nil && !apierrors.IsNotFound(err) {
			log.Errorf("problem removing target namespace [%s] :  %v", targetNamespaceName, err)
		}
		log.Info("Removing the ServiceCatalogAPIServer CR")
		err = operatorConfigClient.ServiceCatalogAPIServers().Delete("cluster", &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("ServiceCatalogAPIServer cr deletion failed: %v", err)
		} else {
			log.Info("ServiceCatalogAPIServer cr removed successfully.")
		}
	default:
		log.Error("Unknown managementState")
	}
}
