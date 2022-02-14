# -*- mode: Python -*-

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/manager main.go'
image = 'quay.io/' + os.environ['repo'] + '/vault-config-operator'

# Go Build
local_resource(
  'vault-config-operator-compile',
  compile_cmd,
  deps=['./main.go','./api','./controllers']
)

# Container Build
custom_build(
  image,
  'podman build -t $EXPECTED_REF --ignorefile ci.Dockerfile.dockerignore -f ./ci.Dockerfile .  && podman push $EXPECTED_REF $EXPECTED_REF',
  entrypoint=['/manager'],
  deps=['./bin'],
  live_update=[
    sync('./bin/manager',"/manager"),
  ],
  skips_local_docker=True,
)

# Manifest Generation
local_resource(
  'vault-config-operator-manifests',
  'make manifests',
  deps=['./bin']
)

allow_k8s_contexts(k8s_context())

# Local Dev
watch_settings(ignore="./config/local-development/tilt/*")
local('envsubst < ./config/local-development/tilt/env-replace-image.yaml > ./config/local-development/tilt/replace-image.yaml', echo_off=True)

k8s_yaml(kustomize('./config/local-development/tilt'))
k8s_resource('vault-config-operator-controller-manager',
  resource_deps=['vault-config-operator-compile', 'vault-config-operator-manifests'])