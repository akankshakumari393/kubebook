### Create kind cluster

```
# create kind cluster with latest kubernetes version
kind create cluster 
```

### Install Volume snapshot CRDS

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-6.1/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-6.1/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-6.1/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml

```

# Create snapshot controller

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-6.1/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/release-6.1/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml

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
[Install snapshot CRDS](https://docs.trilio.io/kubernetes/appendix/csi-drivers/installing-volumesnapshot-crds)
[Official docs](https://github.com/kubernetes-csi/csi-driver-host-path/blob/master/docs/deploy-1.17-and-later.md)
[https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/](https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/)