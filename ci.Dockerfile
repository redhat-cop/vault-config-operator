FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:d91be7cea9f03a757d69ad7fcdfcd7849dba820110e7980d5e2a1f46ed06ea3b
WORKDIR /
COPY bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
