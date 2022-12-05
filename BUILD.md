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













### Implementing v1 version of Resources

```
kubebuilder create api --group batch --version v1 --kind Backup
```

this fails with error
```
failed to create API: unable to inject the resource to "base.go.kubebuilder.io/v3": multiple groups are not allowed by default, to enable multi-group visit https://kubebuilder.io/migration/multi-group.html
```

Make the project directory structure to be compatible to support multiple version
```
mkdir apis/backstore
mkdir -p apis/backstore
mv api/* apis/backstore
rm -rf api/ 
mkdir controllers/backstore
mv controllers/* controllers/backstore/
```

Resolve all the dependencies

kubebuilder create api --group backstore --version v1 --kind Backup [create only resource and not controller]

kubebuilder create api --group backstore --version v1 --kind Restore

Make all changes in v1 Backup and Restore Type

Run `make install`
This would modify the CRDs




### Reference
[Kubebuilder CRD tags](https://book.kubebuilder.io/reference/markers/crd.html)
[kubernetes-controllers-at-scale-clients-caches-conflicts-patches-explained](https://medium.com/@timebertt/kubernetes-controllers-at-scale-clients-caches-conflicts-patches-explained-aa0f7a8b4332)
[https://cloud.redhat.com/blog/kubernetes-operators-best-practices](https://cloud.redhat.com/blog/kubernetes-operators-best-practices)
[https://developer.ibm.com/tutorials/create-a-conversion-webhook-with-the-operator-sdk/](https://developer.ibm.com/tutorials/create-a-conversion-webhook-with-the-operator-sdk/)