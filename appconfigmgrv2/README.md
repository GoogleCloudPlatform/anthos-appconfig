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

