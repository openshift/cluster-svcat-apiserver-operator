# cluster-svcat-apiserver-operator
The openshift-svcat-apiserver-operator installs and maintains openshift/service-catalog on a cluster


## for immediate use
1. Use openshift/installer to install a cluster
2. The following install you will be using the latest svcat-apiserver operator pushed to jboyd01 docker repo - official builds & repo are not setup yet.
3. oc apply the following:
```
 oc apply -f manifests/0000_61_openshift-apiserver-operator_00_namespace.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_03_configmap.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_04_roles.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_05_serviceaccount.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_06_service.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_07_clusteroperator.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_08_deployment.yaml
```

Watch for service catalog apiservers to come up in the kube-service-catalog namespace.

verification:
```
oc get clusteroperators openshift-svcat-apiserver
NAME                        VERSION   AVAILABLE   PROGRESSING   FAILING   SINCE
openshift-svcat-apiserver             True        False         False     10m
```
Review operator deployment events, ensure its not looping:
```
oc describe deployment openshift-svcat-apiserver-operator -n openshift-svcat-apiserver-operator
```


## Recommended development flow
1. Use openshift/installer to install a cluster
2. `make images`
3. `docker tag openshift/origin-cluster-svcat-apiserver-operator:latest <yourdockerhubid>/origin-cluster-svcat-apiserver-operator:latest`
4. `docker push <yourdockerhubid>/origin-cluster-svcat-apiserver-operator:latest`
5. edit manifests/0000_61_openshift-apiserver-operator_07_deployment.yaml update the containers/image to `<yourdockerhubid>/origin-cluster-svcat-apiserver-operator:latest` and update the pull policy to `Always`
6.  oc apply the following:
```
 oc apply -f manifests/0000_61_openshift-apiserver-operator_00_namespace.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_03_configmap.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_04_roles.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_05_serviceaccount.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_06_service.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_07_clusteroperator.yaml
 oc apply -f manifests/0000_61_openshift-apiserver-operator_08_deployment.yaml
```


