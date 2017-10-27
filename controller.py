"""A controller for turning WarmImage CRDs into real things."""

import hashlib
import httplib
import json
from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
import logging
import os
import time

DOMAIN = "mattmoor.io"
VERSION = "v1"
PLURAL = "warmimages"

class WarmImage(object):
    def __init__(self, obj):
        self._obj = obj
        self._metadata = obj["metadata"]
        self._spec = obj["spec"]
        self._image = self._spec["image"]
        self._secrets = self._spec.get("imagePullSecrets")
        self._version = hashlib.sha1(json.dumps(
            self._spec, sort_keys=True)).hexdigest()

    def name(self):
        return self.crd_name() + "-" + self.version()

    def crd_name(self):
        return self._metadata["name"]

    def version(self):
        return self._version

    # Returns a selector that matches any version of
    # a DaemonSet for this WarmImage.
    def any_versions(self):
        return "name=" + self.crd_name()

    # Returns a selector that matches versions other
    # than the current version for this WarmImage.
    def other_versions(self):
        return ",".join([
            self.any_versions(),
            "version!=" + self.version(),
        ])

    def image(self):
        return self._image

    def __str__(self):
        return json.dumps(self._obj, indent=1)

    def daemonset(self, owner_refs):
        return {
            "apiVersion": "extensions/v1beta1",
            "kind": "DaemonSet",
            "metadata": {
                "name": self.name(),
                "ownerReferences": owner_refs,
                "labels": {
                    "name": self.crd_name(),
                    "version": self.version(),
                },
            },
            "spec": {
                "template": {
                    "metadata": {
                        "labels": {
                            "name": self.crd_name(),
                            "version": self.version(),
                        },
                    },
                    "spec": {
                        "containers": [{
                            "name": "the-image",
                            "image": self.image(),
                            # TODO(mattmoor): Do something better than this.
                            "command": ["/bin/sh"],
                            "args": ["-c", "sleep 10000000000"],
                        }],
                        "imagePullSecrets": self._secrets,
                    },
                },
            },
        }


def main():
    config.load_incluster_config()

    apps_beta1 = client.AppsV1beta1Api()
    crds = client.CustomObjectsApi()

    # Create DaemonSets within our own namespace,
    # owned by us (so they go away if we do).
    namespace = os.environ["MY_NAMESPACE"]
    owner = apps_beta1.read_namespaced_deployment(os.environ["OWNER_NAME"], namespace)

    # Define our OwnerReference that we will add to the metadata of
    # objects we create so that they are garbage collected when this
    # controller is deleted.
    controller_ref = {
        "apiVersion": owner.api_version,
        "blockOwnerDeletion": True,
        "controller": True,
        "kind": owner.kind,
        "name": os.environ["OWNER_NAME"],
        "uid": owner.metadata.uid,
    }

    ext_beta1 = client.ExtensionsV1beta1Api()

    def create_meta(warmimage):
        ds = ext_beta1.create_namespaced_daemon_set(namespace, warmimage.daemonset([controller_ref]))
        logging.error("Created the DaemonSet: %s", ds)

    def update_meta(warmimage):
        try:
            # Start warming up the images.
            create_meta(warmimage)
        except ApiException as e:
            if e.status != httplib.CONFLICT:
                raise e

        # Tear down any versions that shouldn't exist.
        delete_meta(warmimage.other_versions())

    def delete_meta(selector):
        for ds in ext_beta1.list_namespaced_daemon_set(
                namespace, label_selector=selector).items:
            ext_beta1.delete_namespaced_daemon_set(
                ds.metadata.name, namespace, body=client.V1DeleteOptions(
                    propagation_policy='Foreground', grace_period_seconds=5))
            logging.error("Deleted the DaemonSet for: %s", str(warmimage))

    def process_meta(t, warmimage, obj):
        if t == "DELETED":
            delete_meta(warmimage.any_versions())
        elif t in ["MODIFIED", "ADDED"]:
            update_meta(warmimage)
        else:
            logging.error("Unrecognized type: %s", t)

    resource_version = ""
    while True:
        stream = watch.Watch().stream(crds.list_cluster_custom_object,
                                      DOMAIN, VERSION, PLURAL,
                                      resource_version=resource_version)
        for event in stream:
            try:
                t = event["type"]
                obj = event["object"]
                warmimage = WarmImage(obj)
                process_meta(t, warmimage, obj)

                # Configure where to resume streaming.
                metadata = obj.get("metadata")
                if metadata:
                    resource_version = metadata["resourceVersion"]
            except:
                logging.exception("Error handling event")

if __name__ == "__main__":
    main()
