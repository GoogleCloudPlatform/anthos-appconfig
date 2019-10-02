# Anthos AppConfig CRD

## Testing

### Integration & Unit

Running unit and integration tests:

```
make test
```

The output of will show test coverage.

Integration tests are ran by spinning up the kubernetes control plane and asserting that expected resources are created. Test isolation is accomplished by spinning up a reconciler and creating an instance of the CRD at the beginning of each test case:

```go
func TestSomething(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, true)
	defer cleanup()

	# Assert that expected resources are created.
}
```

### End-to-End

End-to-end tests are defined at `$REPO_ROOT/tests`. They are written in python and executed on GCP via a cloudbuild job.

## Environment (go 1.12)

```bash
cd ./appconfigmgrv2
export KUBECONFIG= # for make commands that do deployment during testing locally
export GO111MODULE=on
export GOPATH= # - e.g. /Users/joseret/go112
export PATH=$PATH:/usr/local/kubebuilder2/bin # add kubebuilder
```

1. rm go.mod and go.sum
2. go mod init
3. go get sigs.k8s.io/controller-runtime@v0.2.0-beta.2