package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "gazelle", "go_binary", "go_library", "go_prefix")
load("@k8s_object//:defaults.bzl", "k8s_object")

k8s_object(
    name = "controller",
    images = {
        "github.com/mattmoor/warm-image/cmd/controller": "//cmd/controller:image",
        "github.com/mattmoor/warm-image/cmd/sleeper": "//cmd/sleeper:image",
    },
    template = "controller.yaml",
)

k8s_object(
    name = "namespace",
    template = "a-namespace.yaml",
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
