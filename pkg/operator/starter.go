package operator

import (
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"

	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/operator/operatorclient"
	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/operator/resourcesynccontroller"
	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/operator/workloadcontroller"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorv1client "github.com/openshift/client-go/operator/clientset/versioned"
	operatorv1informers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	apiregistrationclient "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	apiregistrationinformers "k8s.io/kube-aggregator/pkg/client/informers/externalversions"
)

func RunOperator(ctx *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(ctx.ProtoKubeConfig)
	if err != nil {
		return err
	}
	apiregistrationv1Client, err := apiregistrationclient.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	operatorConfigClient, err := operatorv1client.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	dynamicClient, err := dynamic.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	configClient, err := configv1client.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}

	v1helpers.EnsureOperatorConfigExists(
		dynamicClient,
		v311_00_assets.MustAsset("v3.11.0/openshift-svcat-apiserver/operator-config.yaml"),
		schema.GroupVersionResource{Group: operatorv1.GroupName, Version: operatorv1.GroupVersion.Version, Resource: "openshiftapiservers"},
	)

	operatorConfigInformers := operatorv1informers.NewSharedInformerFactory(operatorConfigClient, 10*time.Minute)
	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(kubeClient,
		"",
		operatorclient.UserSpecifiedGlobalConfigNamespace,
		operatorclient.MachineSpecifiedGlobalConfigNamespace,
		operatorclient.KubeAPIServerNamespaceName,
		operatorclient.OperatorNamespace,
		operatorclient.TargetNamespaceName,
		"kube-system",
	)
	apiregistrationInformers := apiregistrationinformers.NewSharedInformerFactory(apiregistrationv1Client, 10*time.Minute)
	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)

	operatorClient := &operatorclient.OperatorClient{
		Informers: operatorConfigInformers,
		Client:    operatorConfigClient.OperatorV1(),
	}

	resourceSyncController, err := resourcesynccontroller.NewResourceSyncController(
		operatorClient,
		kubeInformersForNamespaces,
		v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformersForNamespaces),
		v1helpers.CachedSecretGetter(kubeClient.CoreV1(), kubeInformersForNamespaces),
		ctx.EventRecorder,
	)
	if err != nil {
		return err
	}

	workloadController := workloadcontroller.NewWorkloadController(
		os.Getenv("IMAGE"),
		operatorConfigInformers.Operator().V1().ServiceCatalogAPIServers(),
		kubeInformersForNamespaces.InformersFor(operatorclient.TargetNamespaceName),
		kubeInformersForNamespaces.InformersFor(operatorclient.EtcdNamespaceName),
		kubeInformersForNamespaces.InformersFor(operatorclient.KubeAPIServerNamespaceName),
		kubeInformersForNamespaces.InformersFor(operatorclient.UserSpecifiedGlobalConfigNamespace),
		apiregistrationInformers,
		configInformers,
		operatorConfigClient.OperatorV1(),
		configClient.ConfigV1(),
		kubeClient,
		apiregistrationv1Client.ApiregistrationV1(),
		ctx.EventRecorder,
	)
	finalizerController := NewFinalizerController(
		kubeInformersForNamespaces.InformersFor(operatorclient.TargetNamespaceName),
		kubeClient.CoreV1(),
		ctx.EventRecorder,
	)

	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		"service-catalog-apiserver",
		append(
			[]configv1.ObjectReference{
				//TODO: this should be a service catalog api server config map
				{Group: "operator.openshift.io", Resource: "openshiftapiservers", Name: "svcat"},
				{Resource: "namespaces", Name: operatorclient.UserSpecifiedGlobalConfigNamespace},
				{Resource: "namespaces", Name: operatorclient.MachineSpecifiedGlobalConfigNamespace},
				{Resource: "namespaces", Name: operatorclient.OperatorNamespace},
				{Resource: "namespaces", Name: operatorclient.TargetNamespaceName},
			},
			workloadcontroller.APIServiceReferences()...,
		),
		configClient.ConfigV1(),
		operatorClient,
		status.NewVersionGetter(),
		ctx.EventRecorder,
	)

	// make sure our Operator CR exists before proceeding
	glog.Info("waiting for `cluster` ServiceCatalogAPIServer resource to exist")
	err = wait.PollImmediateInfinite(10*time.Second, func() (bool, error) {
		var err error
		_, err = operatorConfigClient.OperatorV1().ServiceCatalogAPIServers().Get("cluster", metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	if err != nil {
		glog.Info("error locating svcat resource: %v", err)
		return err
	}

	operatorConfigInformers.Start(ctx.Done())
	kubeInformersForNamespaces.Start(ctx.Done())
	apiregistrationInformers.Start(ctx.Done())
	configInformers.Start(ctx.Done())

	go workloadController.Run(1, ctx.Done())
	go clusterOperatorStatus.Run(1, ctx.Done())
	go finalizerController.Run(1, ctx.Done())
	go resourceSyncController.Run(1, ctx.Done())

	<-ctx.Done()
	return fmt.Errorf("stopped")
}
