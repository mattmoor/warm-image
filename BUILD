package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "gazelle", "go_binary", "go_library", "go_prefix")

go_prefix("github.com/mattmoor/warm-image")

gazelle(
    name = "gazelle",
    external = "vendored",
)

load("@k8s_object//:defaults.bzl", "k8s_object")

k8s_object(
    name = "controller",
    images = {
        "warmimage-controller:latest": "//cmd/controller:image",
        "sleeper:latest": "//cmd/sleeper:image",
    },
    template = "controller.yaml",
)

k8s_object(
    name = "namespace",
    template = "namespace.yaml",
)

k8s_object(
    name = "serviceaccount",
    template = "serviceaccount.yaml",
)

k8s_object(
    name = "clusterrolebinding",
    template = "clusterrolebinding.yaml",
)

k8s_object(
    name = "warmimage",
    template = "warmimage.yaml",
)

load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")

k8s_objects(
    name = "authz",
    objects = [
        ":serviceaccount",
        ":clusterrolebinding",
    ],
)

k8s_objects(
    name = "everything",
    objects = [
        ":namespace",
        ":authz",
        ":warmimage",
        ":controller",
    ],
)

k8s_object(
    name = "example-warmimage",
    template = ":example-warmimage.yaml",
)

k8s_object(
    name = "release",
    template = "release.yaml",
)
