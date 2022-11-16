### create kind cluster

```
# create kind cluster with latest kubernetes version
kind create cluster 
```

### install Volume snapshot CRDS

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-5.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-5.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-5.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml

```
### Install CSI driver hostpath

```
git clone https://github.com/kubernetes-csi/csi-driver-host-path.git
cd csi-driver-host-path
./deploy/kubernetes-latest/deploy.sh
```

### create storage class and make it default

```
kubectl create -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-host-path/master/examples/csi-storageclass.yaml
storageclass.storage.k8s.io/csi-hostpath-sc created


kubectl patch storageclass standard \
    -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'

kubectl patch storageclass csi-hostpath-sc \
    -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

```












### Technical Reference
[install snapshot CRDS](https://docs.trilio.io/kubernetes/appendix/csi-drivers/installing-volumesnapshot-crds)