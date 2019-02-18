# cluster-svcat-apiserver-operator
The cluster-svcat-apiserver-operator installs and maintains openshift/service-catalog on a cluster.  This operator only deals with the API Server portion of Service Catalog; see also the `cluster-svcat-controller-manager-operator`.

Note that the manifests do not create the Cluster Operator or the ServiceCatalogAPIServer custom resource.  While the CVO installs the Service Catalog operators, we don't want Service Catalog installed by default.  The cluster admin must create the ServiceCatalogAPIServer CR to cause the operator to perform the installation ([see below](#Trigger-installation-of-Service-Catalog-API-Server))

Once the operator detects the CR it will create the Service Catalog Cluster Operator resource and proceed with reconciling the Service Catalog API Server deployment.

## Deployment the operator prior to CVO integration
1. Use openshift/installer to install a cluster.  Skip to step 6 if you want to use pre-built operator images.
2. `make images`
3. `docker tag openshift/origin-cluster-svcat-apiserver-operator:latest <yourdockerhubid>/origin-cluster-svcat-apiserver-operator:latest`
4. `docker push <yourdockerhubid>/origin-cluster-svcat-apiserver-operator:latest`
5. edit manifests/0000_61_openshift-svcat-apiserver-operator_08_deployment.yaml and update the containers/image to `<yourdockerhubid>/origin-cluster-svcat-apiserver-operator:latest` and set the pull policy to `Always`
6.  `oc apply -f manifests`

This will cause the creation of the cluster-svcat-apiserver-operator deployment 
and associated resources.  The operator waits for creation of the `ServiceCatalogAPIServer`
custom resource before doing any real work including creating the Cluster Operator `openshift-svcat-apiserver`.  

## Trigger installation of Service Catalog API Server
Create the `ServiceCatalogAPIServer` CR to trigger the installation of Service Catalog:
```
$ cat <<'EOF' | oc create -f -
apiVersion: operator.openshift.io/v1
kind: ServiceCatalogAPIServer
metadata:
  name: cluster
spec:
  managementState: Managed
EOF
```
Once the cluster `ServiceCatalogAPIServer` is found to exist and have a `managementState` of `Managed` the operator will create necessary resources in the
`kube-service-catalog` namespace for deploying the Service Catalog API Server.

Watch for service catalog apiservers to come up in the kube-service-catalog namespace.

## Verification & debugging
Nothing happens without the CR:
```
$ oc get servicecatalogapiservers
NAME      AGE
cluster     10m
```
If the state is `Managed` the operator will install Service Catalog API Server.  You can remove the deployment by setting the state to `Removed`.  

Once the CR is created the operator should create a new ClusterOperator resource:
```
oc get clusteroperator openshift-svcat-apiserver
NAME                        VERSION   AVAILABLE   PROGRESSING   FAILING   SINCE
openshift-svcat-apiserver             True        False         False     1m
```
Review operator pod logs from the `openshift-svcat-apiserver` namespace to see details of the operator processing.


The operator deployment events will give you an overview of what it's done.  Ensure its not looping & review the events:
```
$ oc describe deployment openshift-svcat-apiserver-operator -n openshift-svcat-apiserver-operator
```





