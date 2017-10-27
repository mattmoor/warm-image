"""A controller for turning WarmImage CRDs into real things."""

import json
from kubernetes import client, config, watch
import logging
import os

DOMAIN = "mattmoor.io"
VERSION = "v1"
PLURAL = "warmimages"


class WarmImage(object):
    def __init__(self, obj):
        self._obj = obj
        self._metadata = obj["metadata"]
        self._image = obj["spec"]["image"]
        # TODO(mattmoor): Support optional pull secrets.

    def name(self):
        return self._metadata["name"]

    def image(self):
        return self._image

    def __str__(self):
        return json.dumps(self._obj, indent=1)

    def daemonset(self, owner_refs):
        return {
            "apiVersion": "extensions/v1beta1",
            "kind": "DaemonSet",
            "metadata": {
                # TODO(mattmoor): Does this disambiguate enough?
                "name": self.name(),
                "ownerReferences": owner_refs,
            },
            "spec": {
                "template": {
                    "metadata": {
                        "labels": {
                            "name": self.name(),
                        },
                    },
                    "spec": {
                        "containers": [{
                            "name": "the-image",
                            "image": self.image(),
                            # TODO(mattmoor): Do something better than this.
                            "command": ["/bin/sh"],
                            "args": ["-c", "sleep 10000000000"],
                            # TODO(mattmoor): Support pull secrets.
                        }],
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
        logging.error("TODO Update the DaemonSet for: %s", str(warmimage))

    def delete_meta(warmimage):
        ext_beta1.delete_namespaced_daemon_set(
            warmimage.name(), namespace, body=client.V1DeleteOptions(
                propagation_policy='Foreground', grace_period_seconds=5))
        logging.error("Deleted the DaemonSet for: %s", str(warmimage))

    def process_meta(t, warmimage, obj):
        if t == "DELETED":
            delete_meta(warmimage)
        elif t == "MODIFIED":
            update_meta(warmimage)
        elif t == "ADDED":
            create_meta(warmimage)
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
