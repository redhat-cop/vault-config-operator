# Default values for helm-try.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

# Global parameters
global:
  # global.imageRegistry -- Global Docker image registry (overrides image.registry)
  imageRegistry: ""

replicaCount: 1

image:
  # image.registry -- Image registry (overridden by global.imageRegistry if set)
  registry: ${image_registry}
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
podLabels: {}

podSecurityContext:
  runAsNonRoot: true

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - "ALL"

resources:
  requests:
    cpu: 100m
    memory: 250Mi

nodeSelector: {}

tolerations: []

affinity: {}

# Set to true to create a ServiceMonitor resource (requires Prometheus Operator CRDs).
enableMonitoring: false
metricsSecure: true
enableCertManager: false

certificate:
  issuerRef:
    name: selfsigned-issuer
    kind: Issuer
    # group: cert-manager.io
