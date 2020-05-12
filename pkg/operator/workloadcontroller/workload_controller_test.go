package workloadcontroller

import (
	"strings"
	"testing"

	operatorv1 "github.com/openshift/api/operator/v1"
	operatorfake "github.com/openshift/client-go/operator/clientset/versioned/fake"
	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/operator/operatorclient"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/status"
	operatorv1helpers "github.com/openshift/library-go/pkg/operator/v1helpers"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	kubeaggregatorfake "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/fake"
)

func TestUpgradeable(t *testing.T) {

	testCases := []struct {
		name            string
		managementState operatorv1.ManagementState
		expectedStatus  operatorv1.ConditionStatus
		expectedMessage string
		substring       bool
	}{
		{
			name:            "in managed state, upgradeable should be false",
			managementState: operatorv1.Managed,
			expectedStatus:  operatorv1.ConditionFalse,
			substring:       true,
			expectedMessage: "https://docs.openshift.com/container-platform/4.4/applications/service_brokers/installing-service-catalog.html",
		},
		{
			name:            "in unmanaged state, upgradeable should be true",
			managementState: operatorv1.Unmanaged,
			expectedStatus:  operatorv1.ConditionTrue,
			substring:       true,
			expectedMessage: "unmanaged state, upgrades are possible",
		},
		{
			name:            "in removed state, upgradeable should be true",
			managementState: operatorv1.Removed,
			expectedStatus:  operatorv1.ConditionTrue,
			substring:       true,
			expectedMessage: "removed state, upgrades are possible",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// create kube client
			kubeClient := fake.NewSimpleClientset(
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "apiserver",
						Namespace:  operatorclient.TargetNamespaceName,
						Generation: 100,
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:    100,
						ObservedGeneration: 100,
					},
				})

			// create operator config
			operatorConfig := &operatorv1.ServiceCatalogAPIServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "cluster",
					Generation: 100,
				},
				Spec: operatorv1.ServiceCatalogAPIServerSpec{
					OperatorSpec: operatorv1.OperatorSpec{
						ManagementState: tc.managementState,
					},
				},
				Status: operatorv1.ServiceCatalogAPIServerStatus{
					OperatorStatus: operatorv1.OperatorStatus{
						ObservedGeneration: 100,
					},
				},
			}

			// create apiserver client and required clients
			apiServiceOperatorClient := operatorfake.NewSimpleClientset(operatorConfig)
			kubeAggregatorClient := kubeaggregatorfake.NewSimpleClientset()

			// create dynamic client
			dynamicScheme := runtime.NewScheme()
			dynamicScheme.AddKnownTypeWithName(schema.GroupVersionKind{
				Group:   "monitoring.coreos.com",
				Version: "v1",
				Kind:    "ServiceMonitor"},
				&unstructured.Unstructured{})
			dynamicClient := dynamicfake.NewSimpleDynamicClient(dynamicScheme)

			// Finally create the operator we need to test
			operator := ServiceCatalogAPIServerOperator{
				kubeClient:              kubeClient,
				eventRecorder:           events.NewInMemoryRecorder(""),
				operatorConfigClient:    apiServiceOperatorClient.OperatorV1(),
				dynamicClient:           dynamicClient,
				apiregistrationv1Client: kubeAggregatorClient.ApiregistrationV1(),
				versionRecorder:         status.NewVersionGetter(),
			}

			// Test the sync method
			err := operator.sync()
			if err != nil {
				t.Fatal("Unexpected error running sync")
			}

			// verify results
			result, err := apiServiceOperatorClient.OperatorV1().
				ServiceCatalogAPIServers().Get("cluster", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			condition := operatorv1helpers.FindOperatorCondition(
				result.Status.Conditions, operatorv1.OperatorStatusTypeUpgradeable)
			if condition == nil {
				t.Fatalf("No %v condition found.", operatorv1.OperatorStatusTypeUpgradeable)
			}

			// verify the status
			if condition.Status != tc.expectedStatus {
				t.Errorf("expected %v but received %v", tc.expectedStatus, condition.Status)
			}

			// verify the status messages
			if tc.substring {
				if !strings.Contains(condition.Message, tc.expectedMessage) {
					t.Fatalf("expected to find %v in the message: %v", tc.expectedMessage, condition.Message)
				}
			} else {
				if condition.Message != tc.expectedMessage {
					t.Fatalf("expected %v to match message:%v", tc.expectedMessage, condition.Message)
				}
			}
		})
	}
}
