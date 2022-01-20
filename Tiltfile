# -*- mode: Python -*-

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/manager main.go'
image = 'quay.io/' + os.environ['repo'] + '/vault-config-operator'

local_resource(
  'vault-config-operator-compile',
  compile_cmd,
  deps=['./main.go','./api','./controllers'])


custom_build(
  image,
  'podman build -t $EXPECTED_REF --ignorefile ci.dockerignore -f ./ci.Dockerfile .  && podman push $EXPECTED_REF $EXPECTED_REF',
  entrypoint=['/manager'],
  deps=['./bin'],
  live_update=[
    sync('./bin/manager',"/manager"),
  ],
  skips_local_docker=True,
)

allow_k8s_contexts(k8s_context())
k8s_yaml(kustomize('./config/local-development/tilt'))
k8s_resource('vault-config-operator-controller-manager',resource_deps=['vault-config-operator-compile'])