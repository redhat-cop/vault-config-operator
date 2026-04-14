FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:fe688da81a696387ca53a4c19231e99289591f990c904ef913c51b6e87d4e4df
WORKDIR /
COPY bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
