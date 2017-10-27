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


def main():
    config.load_incluster_config()

    api_ext = client.ApiextensionsV1beta1Api()
    apps = client.AppsV1beta1Api()
    crds = client.CustomObjectsApi()

    # Create DaemonSets within our own namespace,
    # owned by us (so they go away if we do).
    namespace = os.environ["MY_NAMESPACE"]
    owner = apps.read_namespaced_deployment(os.environ["OWNER_NAME"], namespace)

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

    def create_meta(warmimage):
        logging.error("Create the DaemonSet for: %s", str(warmimage))

    def update_meta(warmimage):
        logging.error("Update the DaemonSet for: %s", str(warmimage))

    def delete_meta(warmimage):
        logging.error("Delete the DaemonSet for: %s", str(warmimage))

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
