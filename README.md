# Kubernetes `WarmImage` CRD.

## Overview

This branch defines a Proof-of-Concept implementation of the Knative Image cache
resource that prefetches it onto every node in your cluster using a DaemonSet.

## Cluster Setup

To install this onto your cluster, you may simply run:
```shell
# Install the CRD and Controller.
curl https://raw.githubusercontent.com/mattmoor/warm-image/poc-cache/release.yaml \
  | kubectl apply -f -
```

Alternately you may `git clone` this repository and run:
```shell
# Install the CRD and Controller.
kubectl apply -f release.yaml
```

### Uninstall

Simply use the same command you used to install, but with `kubectl delete` instead of `kubectl apply`.

## Usage

### Specification

The specification for an image to "warm up" looks like:
```yaml
apiVersion: caching.internal.knative.dev/v1alpha1
kind: Image
metadata:
  name: example-warmimage
spec:
  image: gcr.io/google-appengine/debian8:latest
  # Optionally:
  # imagePullSecrets:
  # - name: foo
  # Or:
  # serviceAccountName: bar
```

### Creation

With the above in `foo.yaml`, you would install the image with:
```shell
kubectl apply -f foo.yaml
```

### Listing

You can see what images are "warm" via:
```shell
$ kubectl get images
NAME                KIND
example-warmimage   Image.v1alpha1.caching.internal.knative.dev
```

### Updating

You can upgrade `foo.yaml` to `debian9` and run:
```shell
kubectl apply -f foo.yaml
```

### Removing

You can remove a warmed image via:
```shell
kubectl delete image example-warmimage
```

