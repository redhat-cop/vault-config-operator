# Default values for helm-try.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

image:
  repository: ${image_repo}
  pullPolicy: IfNotPresent
  # Overrides the image tag whose default is the chart appVersion.
  tag: ${version}

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""
env: []
args: []
volumes: []
volumeMounts: []
podAnnotations: {}

resources:
  requests:
    cpu: 100m
    memory: 250Mi

nodeSelector: {}

tolerations: []

affinity: {}

kube_rbac_proxy:
  securePort: 8443
  image:
    repository: quay.io/redhat-cop/kube-rbac-proxy
    pullPolicy: IfNotPresent
    tag: v0.11.0
  resources:
    limits:
      cpu: 500m
      memory: 128Mi
    requests:
      cpu: 5m
      memory: 64Mi

metricsPort: 8080
healthProbePort: 8081
hostNetwork: false
enableMonitoring: true
enableCertManager: false
