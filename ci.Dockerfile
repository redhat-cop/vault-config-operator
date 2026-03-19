FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:83006d535923fcf1345067873524a3980316f51794f01d8655be55d6e9387183
WORKDIR /
COPY bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
