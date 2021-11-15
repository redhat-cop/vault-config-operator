#/bin/bash
# 
# Used by Tiltfile and make helmchart to dinamically set the controller image.
#

image="quay.io/$repo/vault-config-operator"
cd ./config/manager && kustomize edit set image controller=$image:latest
