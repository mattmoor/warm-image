package(default_visibility = ["//visibility:public"])

load("@warmimage_pip//:requirements.bzl", "requirement")
load("@io_bazel_rules_docker//python:image.bzl", "py_image")

py_image(
    name = "controller",
    srcs = ["controller.py"],
    main = "controller.py",
    deps = [requirement("kubernetes")],
)

load("@k8s_crd//:defaults.bzl", "k8s_crd")

k8s_crd(
    name = "warmimage-crd",
    template = ":warmimage.yaml",
)

load("@k8s_deployment//:defaults.bzl", "k8s_deployment")

k8s_deployment(
    name = "deployment",
    images = {
        "gcr.io/mattmoor-public/warm-image/controller:latest": ":controller",
    },
    template = ":controller.yaml",
)

load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")

k8s_objects(
    name = "everything",
    objects = [
        ":warmimage-crd",
        ":deployment",
    ],
)

load("@k8s_object//:defaults.bzl", "k8s_object")

k8s_object(
    name = "example-warmimage-crd",
    template = ":example-warmimage.yaml",
)
