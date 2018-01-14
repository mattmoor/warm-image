# Kubernetes `WarmImage` CRD.

## Overview

This repository defines a Kubernetes "Custom Resource Definition" (CRD) called
`WarmImage`.  The `WarmImage` CRD takes an image reference (with optional
secrets) and prefetches it onto every node in your cluster.

## Cluster Setup

**It is recommended that folks install this into its own namespace.**

To install this custom resource onto your cluster, you may simply run:
```shell
# Install the CRD and Controller.
curl https://raw.githubusercontent.com/mattmoor/warm-image/master/release.yaml \
  | kubectl create -f -
```

Alternately you may `git clone` this repository and run:
```shell
# Install the CRD and Controller.
# TODO(mattmoor): NEEDS UPDATE
kubectl create -f release.yaml
```

### Uninstall

Simply use the same command you used to install, but with `kubectl delete` instead of `kubectl create`.

TODO(mattmoor): NEEDS UPDATE This will result in the 404 deleting the controller.

## Usage

### Specification

The specification for an image to "warm up" looks like:
```yaml
apiVersion: mattmoor.io/v2
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
example-warmimage   WarmImage.v2.mattmoor.io
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

