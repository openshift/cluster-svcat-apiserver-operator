# cluster-svcat-apiserver-operator
The cluster-svcat-apiserver-operator installs and maintains a singleton instance of the OpenShift Service Catalog on a cluster.  Service Catalog is actually comprised of an aggregated API server and a controller manager; this operator only deals with the API Server portion of Service Catalog.  See the [cluster-svcat-controller-manager-operator](https://github.com/openshift/cluster-svcat-controller-manager-operator) for the operator responsible for the controller manager component of Service Catalog.

It should be noted this repo was initially copied from the [OpenShift API Server Operator](https://github.com/openshift/cluster-openshift-apiserver-operator) and we generally try to keep it in sync with fixes and updates that are applicable.

[The Cluster Version Operator](https://github.com/openshift/cluster-version-operator) installs cluster operators by collecting the files within each cluster operator's manifest directory, bundling them into a release payload, and then `oc apply`ing them.  Note that unlike most cluster operators, this operator's configuration specifies that the initial management state of the Operator is `Removed`.  That is, the cluster operator is installed and running, but the operand is not.

This operator is installed to the `openshift-service-catalog-apiserver-operator` namespace.  It installs the Service Catalog API Server into the `openshift-service-catalog-apiserver` namespace.  In prior versions, both the Service Catalog API Server and Controller Manager were installed to kube-service-catalog.  This change keeps with how the OpenShift API Server & Controller Manager are managed and makes some aspects of servicability easier.


## Installing Service Catalog
To enable and install Service Catalog, the cluster admin must modify two Service Catalog custom resources and change the `ManagementState` to `Managed`. 
```
$ oc edit ServiceCatalogAPIServer
```
locate the `managementState` and change `Removed` to `Managed`.  Repeat for the controller-manager:
```
$ oc edit ServiceCatalogControllerManager
```
note the latter resource is actually from the [cluster-svcat-controller-manager-operator](https://github.com/openshift/cluster-svcat-controller-manager-operator).  The apiserver operator will see the change in the desired state and create necessary resources in the `openshift-service-catalog-apiserver` namespace for deploying the Service Catalog API Server.


## Verification & debugging
Review the cluster operator status, it should report `Available` if it is in the desired state.  Although a bit contrary to the notion of "available", it should be pointed out that when the managementState is `Removed` and Service Catalog is not installed, the operator should be reporting Available=true because it is in the desired state.
```
$ oc get clusteroperators service-catalog-apiserver
NAME                        VERSION                        AVAILABLE   PROGRESSING   DEGRADED   SINCE
service-catalog-apiserver   4.1.0-0.ci-2019-05-01-061138   True        False         False      3m57s
```
View the operator pod logs:
```
$ oc logs deployment/openshift-service-catalog-apiserver-operator -n openshift-service-catalog-apiserver-operator
```
The events present a good summary of actions the operator has taken to reach the desired state:
```
$ oc get events --sort-by='.lastTimestamp'  -n openshift-service-catalog-apiserver-operator
```

If the state is `Managed` the operator will install Service Catalog API Server.  You can request the Service Catalog deployment to be removed by setting the state to `Removed`.  

## Hacking with your own Operator or Operand
You can make changes to the operator and deploy it to your cluster.  First you disable the CVO so it doesn't overwrite your changes from what is in the release payload:
```
$ oc scale --replicas 0 -n openshift-cluster-version deployments/cluster-version-operator
```
this is a big hammer, you could instead just tell the CVO your operator should be unmanged, see [Setting Objects unmanaged](https://github.com/openshift/cluster-version-operator/blob/master/docs/dev/clusterversion.md#setting-objects-unmanaged)

Build and push your newly built image to a repo:
```
$ make images
$ docker tag registry.svc.ci.openshift.org/ocp/4.2:cluster-svcat-apiserver-operator docker.io/ACCOUNT/ocp-cluster-svcat-apiserver-operator:latest
$ docker push docker.io/ACCOUNT/ocp-cluster-svcat-apiserver-operator:latest
```
and then update the manifest to specify your operator  image:
```
$ oc edit deployment -n openshift-service-catalog-apiserver-operator
```
locate the image and change the image and pull policy:
```
        image: registry.svc.ci.openshift.org/ocp/4.1-2019-05-01-061138@sha256:de5e1c8a2605f75b71705a933c31f4dff3ff1ae860d7a86d771dbe2043a4cea0
        imagePullPolicy: IfNotPresent
```
to
```
        image: docker.io/ACCOUNT/ocp-cluster-svcat-apiserver-operator:latest
        imagePullPolicy: Always
```
This will cause your dev operator image to be pulled down and deployed.  When you want to deploy a newly built image just scale your operator to zero and right back to one:
```
$ oc scale --replicas 0 -n openshift-service-catalog-apiserver-operator deployments/openshift-service-catalog-apiserver-operator
$ oc scale --replicas 1 -n openshift-service-catalog-apiserver-operator deployments/openshift-service-catalog-apiserver-operator
```

If you want your own Service Catalog API Server to be deployed you follow a simlar process but instead update the deployment's IMAGE environment variable:
```
        env:
        - name: IMAGE
          value: registry.svc.ci.openshift.org/ocp/4.1-2019-05-01-061138@sha256:cc22f2af68a261c938fb1ec9a9e94643eba23a3bb8c9e22652139f80ee57681b
```
and change the value to your own repo, something like
```
        env:
        - name: IMAGE
          value: docker.io/ACCOUNT/service-catalog:latest
```

When testing changes to the operator, remember to validate with a simulated fresh cluster install.  This should include:
1) ensure Service Catalog API Server is not installed
2) disabling the cvo (`oc scale --replicas 0 -n openshift-cluster-version deployments/cluster-version-operator`) and operator (`oc scale --replicas 0 -n openshift-service-catalog-apiserver-operator deployments/openshift-service-catalog-apiserver-operator`)
3) deleting the custom resource ServiceCatalogAPIServers (`oc delete servicecatalogapiservers cluster`)
4) creating a fresh CR (`oc apply -f manifests/03_config.cr.yaml`)
5) deleting the ClusterOperator resource (`oc delete clusteroperator service-catalog-apiserver`)
6) Update the `manifests/08_cluster-operator.yaml` to reflect your updated operator image
7) deploy your operator with `$ oc apply -f manifests/08_cluster-operator.yaml`

Make sure the cluster operator comes up and sets the status conditions as follows:
```
$ oc get clusteroperator service-catalog-apiserver -o yaml
apiVersion: config.openshift.io/v1
kind: ClusterOperator
metadata:
.......
spec: {}
  status:
    conditions:
    - lastTransitionTime: 2019-05-17T01:42:33Z
      message: the apiserver is in the desired state (Removed).
      reason: Removed
      status: "True"
      type: Available
    - lastTransitionTime: 2019-05-17T01:41:53Z
      reason: Removed
      status: "False"
      type: Progressing
    - lastTransitionTime: 2019-05-17T01:36:42Z
      reason: Removed
      status: "False"
      type: Degraded
```


## Read about the CVO if you haven't yet
Consider this required reading - its vital to understanding how the operator should work and why:
* https://github.com/openshift/cluster-version-operator#cluster-version-operator-cvo
* https://github.com/openshift/cluster-version-operator/tree/master/docs/dev

## Other development notes
If you make changes to the yaml resources under `bindata` you must run `make update-generated` to update the go source files which are responsible for creating the Service Catalog operand deployment resources.

When picking up new versions of dependencies, use `make update-deps`.  Generally you want to mirror the `glide.yaml` and dependency updates driven from the OpenShift apiserver operator.
