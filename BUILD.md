### Initialize the Project

```
kubebuilder init --domain akankshakumari393.github.io --repo akankshakumari393.github.io/kubebook
```

Create API for BackUp
```
kubebuilder create api --group backstore --version v1beta1 --kind Backup
```

Create API for Restore
```
kubebuilder create api --group backstore --version v1beta1 --kind Restore
```


```
go get github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1
go get github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned
go mod tidy
```


### Local setup with Pod and PVC
```
kubectl create -f config/storage.yaml
```











### Reference
[Kubebuilder CRD tags](https://book.kubebuilder.io/reference/markers/crd.html)
[kubernetes-controllers-at-scale-clients-caches-conflicts-patches-explained](https://medium.com/@timebertt/kubernetes-controllers-at-scale-clients-caches-conflicts-patches-explained-aa0f7a8b4332)
[https://cloud.redhat.com/blog/kubernetes-operators-best-practices](https://cloud.redhat.com/blog/kubernetes-operators-best-practices)