# DF Tool

DF Tool is a small preprocessor for Dockerfiles. It adds a few extra features that I felt was missing from
the standard Dockerfile. DF Tool requires the Docker client to be installed.

## Using

`go get github.com/lfkeitel/dftool`

`dftool [-f Dockerfile.dft] [-t image tag] PATH | URL`

Flags:

- `-f` - The Dockerfile to build. Defaults to "Dockerfile.dft".
- `-t` - The tag given to the generated image. This flag overrides the TAG directive in a Dockerfile.

## FROM

The FROM directive can take a second argument: `FROM alpine:3.5 pull-always`.
`pull-always` will make the build client perform a docker pull on the base image before
starting docker build.

## RUN

The RUN directive is now a bit cleaner. Each command can be placed on a separate line without
needing `&& \` at the end of every line.

```Dockerfile
FROM alpine:3.5

RUN (
    apk update
    apk add nginx \
            vim
    rm -rf /var/cache/apk/*
)
```

And ending backslash is supported for breaking up long commands over separate lines such as apt-get,
or other package installations.

## TAG

TAG is a new directive that specifies the image name and optional tag applied to a build image.
This directive can be overridden with the `-t` flag.

## How does it work?

DF Tool processes a Dockerfile and generates a standard Dockerfile for the standard build tool
but also performs other tasks outside of the normal build command such as pulling images. The
generated Dockerfile is written to the `$CONTEXT_DIR/Dockerfile.tmp`. The generated file is deleted
once the build process is done.