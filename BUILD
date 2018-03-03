package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "gazelle", "go_binary", "go_library", "go_prefix")

go_prefix("github.com/mattmoor/warm-image")

gazelle(
    name = "gazelle",
    external = "vendored",
)

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/mattmoor/warm-image",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/client/clientset/versioned:go_default_library",
        "//pkg/client/informers/externalversions:go_default_library",
        "//pkg/controller:go_default_library",
        "//pkg/controller/warmimage:go_default_library",
        "//pkg/signals:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/client-go/informers:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/tools/clientcmd:go_default_library",
    ],
)

go_binary(
    name = "warm-image",
    embed = [":go_default_library"],
    importpath = "github.com/mattmoor/warm-image",
    pure = "on",
)

load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_image(
    name = "image",
    binary = ":warm-image",
)

load("@k8s_object//:defaults.bzl", "k8s_object")

k8s_object(
    name = "controller",
    images = {
        "warmimage-controller:latest": ":image",
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
