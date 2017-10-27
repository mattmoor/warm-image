# Kubernetes `WarmImage` CRD.

## Overview

This repository defines a Kubernetes "Custom Resource Definition" (CRD) called
`WarmImage`.  The `WarmImage` CRD takes an image reference (with optional
secrets) and prefetches it onto every node in your cluster.

## Cluster Setup

**It is recommended that folks install this into its own namespace.**

To install this custom resource onto your cluster, you may simply run:
```shell
# Create the namespace
kubectl create namespace warm-image
# Install the CRD and Controller.
curl https://raw.githubusercontent.com/mattmoor/warm-image/master/release.yaml \
  | kubectl --namespace=warm-image create -f -
```

Alternately you may `git clone` this repository and run:
```shell
# Create the namespace
kubectl create namespace warm-image
# Install the CRD and Controller.
kubectl  --namespace=warm-image create -f release.yaml
```

### Uninstall

If you have isolated this controller into its own namespace, you can clean
things up by deleting the `warm-image` namespace.  Alternately:
```shell
# This should clean up all of the DaemonSets
kubectl delete deployment warmimage-controller
# This should clean up all of the WarmImages
kubectl delete crd warmimages.mattmoor.io
```

## Usage

### Specification

The specification for an image to "warm up" looks like:
```yaml
apiVersion: "mattmoor.io/v1"
kind: WarmImage
metadata:
  name: example-warmimage
spec:
  image: gcr.io/google-appengine/debian8:latest
  # Optionally:
  # imagePullSecrets: 
  # - name: foo
```

### Creation

With the above in `foo.yaml`, you would install the image with:
```shell
kubectl create -f foo.yaml
```

### Listing

You can see what images are "warm" via:
```shell
$ kubectl get warmimages
NAME                KIND
example-warmimage   WarmImage.v1.mattmoor.io
```

### Updating

You can upgrade `foo.yaml` to `debian9` and run:
```shell
kubectl replace -f foo.yaml
```

### Removing

You can remove a warmed image via:
```shell
kubectl delete warmimage example-warmimage
```

