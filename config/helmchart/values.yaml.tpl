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

enableMonitoring: true
metricsSecure: true
enableCertManager: false
