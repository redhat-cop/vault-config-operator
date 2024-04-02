FROM registry.access.redhat.com/ubi8/ubi-minimal@sha256:f30dbf77b075215f6c827c269c073b5e0973e5cea8dacdf7ecb6a19c868f37f2
WORKDIR /
COPY bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
