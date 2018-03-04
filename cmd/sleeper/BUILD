package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/mattmoor/warm-image/cmd/sleeper",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/golang/glog:go_default_library"],
)

go_binary(
    name = "sleeper",
    embed = [":go_default_library"],
    importpath = "github.com/mattmoor/warm-image/cmd/sleeper",
    pure = "on",
)

load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_image(
    name = "image",
    binary = ":sleeper",
)
