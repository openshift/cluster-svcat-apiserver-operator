package configobservercontroller

import (
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorv1informers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/operator/configobservation"
)

type ConfigObserver struct {
	*configobserver.ConfigObserver
}

// NewConfigObserver initializes a new configuration observer.
func NewConfigObserver(
	operatorClient v1helpers.OperatorClient,
	resourceSyncer resourcesynccontroller.ResourceSyncer,
	operatorConfigInformers operatorv1informers.SharedInformerFactory,
	kubeInformersForEtcdNamespace kubeinformers.SharedInformerFactory,
	configInformers configinformers.SharedInformerFactory,
	eventRecorder events.Recorder,
) *ConfigObserver {
	c := &ConfigObserver{
		ConfigObserver: configobserver.NewConfigObserver(
			operatorClient,
			eventRecorder,
			configobservation.Listers{
				ResourceSync:    resourceSyncer,
				EndpointsLister: kubeInformersForEtcdNamespace.Core().V1().Endpoints().Lister(),
				PreRunCachesSynced: []cache.InformerSynced{
					operatorConfigInformers.Operator().V1().ServiceCatalogAPIServers().Informer().HasSynced,
					kubeInformersForEtcdNamespace.Core().V1().Endpoints().Informer().HasSynced,
				},
			},
		),
	}
	operatorConfigInformers.Operator().V1().ServiceCatalogAPIServers().Informer().AddEventHandler(c.EventHandler())
	kubeInformersForEtcdNamespace.Core().V1().Endpoints().Informer().AddEventHandler(c.EventHandler())

	//TODO delete
	//configInformers.Config().V1().Images().Informer().AddEventHandler(c.EventHandler())
	return c
}
