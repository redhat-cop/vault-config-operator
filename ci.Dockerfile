FROM registry.access.redhat.com/ubi8/ubi-minimal@sha256:8bedbe742f140108897fb3532068e8316900d9814f399d676ac78b46e740e34e
WORKDIR /
COPY bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
