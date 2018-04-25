package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "gazelle", "go_binary", "go_library", "go_prefix")

go_prefix("github.com/mattmoor/warm-image")

gazelle(
    name = "gazelle",
    external = "vendored",
)

load("@k8s_object//:defaults.bzl", "k8s_object")

k8s_object(
    name = "example-warmimage",
    template = ":example-warmimage.yaml",
)

k8s_object(
    name = "release",
    template = "release.yaml",
)
