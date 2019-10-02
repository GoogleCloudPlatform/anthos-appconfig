// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright 2019 Google LLC. This software is provided as-is,
// without warranty or representation for any use or purpose.
//

package controllers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var restConfig *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var scheme = runtime.NewScheme()

// TestMain spins up a local kubernetes control plane to run resource based
// integration tests in.
func TestMain(m *testing.M) {
	if strings.ToLower(os.Getenv("TEST_LOGS")) == "enabled" {
		logf.SetLogger(logf.ZapLogger(false))
	}

	const istioVersion = "1.1.7"
	t := &envtest.Environment{}

	corev1.AddToScheme(scheme)
	netv1.AddToScheme(scheme)
	appconfig.AddToScheme(scheme)
	v1beta1.AddToScheme(scheme)

	var err error
	if restConfig, err = t.Start(); err != nil {
		log.Error(err, "starting test env")
		os.Exit(1)
	}

	// Fail on missing path (unable to do with t.Start).
	// https://github.com/kubernetes-sigs/controller-runtime/issues/481
	if _, err := envtest.InstallCRDs(t.Config, envtest.CRDInstallOptions{
		Paths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "third_party", "istio", "v"+istioVersion, "crds"),
		},
		CRDs:               nil,
		ErrorIfPathMissing: true,
	}); err != nil {
		log.Error(err, "installing crds")
		os.Exit(1)
	}

	code := m.Run()

	t.Stop()
	os.Exit(code)
}

// startTestReconciler starts a manager and returns a stop function.
func startTestReconciler(t *testing.T) (*AppEnvConfigTemplateV2Reconciler, func()) {
	mgr, err := ctrl.NewManager(restConfig, manager.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)

	r := &AppEnvConfigTemplateV2Reconciler{
		Client:         mgr.GetClient(),
		Dynamic:        dynamic.NewForConfigOrDie(mgr.GetConfig()),
		Log:            ctrl.Log.WithName("controllers").WithName("AppEnvConfigTemplateV2"),
		Scheme:         mgr.GetScheme(),
		skipGatekeeper: true,
	}
	require.NoError(t, r.SetupWithManager(mgr))

	return r, startTestManager(t, mgr)
}

// startTestManager starts a manager or fails the given test case.
func startTestManager(t *testing.T, mgr manager.Manager) func() {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		assert.NoError(t, mgr.Start(stop))
	}()

	// Return stop func intended to be deferred.
	return func() {
		close(stop)
		wg.Wait()
	}
}

// createTestNamespace creates a namespace with a name derived from the
// test name.
func createTestNamespace(t *testing.T, f testFeatureFlags) func() {
	c, err := client.New(restConfig, client.Options{Scheme: scheme})
	require.NoError(t, err)

	namespace := testNamespaceName(t)
	name := types.NamespacedName{Name: namespace}

	create := &corev1.Namespace{}
	create.SetName(namespace)
	if f.istio {
		create.SetLabels(map[string]string{"istio-injection": "enabled"})
	}
	require.NoError(t, c.Create(context.TODO(), create))

	found := &corev1.Namespace{}
	require.NoError(t, c.Get(context.TODO(), name, found))

	// Return a cleanup func.
	return func() {
		c.Delete(context.TODO(), create)
	}
}

// testNamespaceName returns the name of the namespace to run the given test in.
// It is derived from the name of the test itself to ensure uniqueness across
// a test suite.
func testNamespaceName(t *testing.T) string {
	return strings.ToLower(t.Name())
}

// testFeatureFlags determine how a namespace is configured for a given test.
type testFeatureFlags struct {
	istio bool
	vault bool
}

// createTestNamespace creates an example instance of an app config.
func createTestInstance(t *testing.T, f testFeatureFlags) (*appconfig.AppEnvConfigTemplateV2, func()) {
	c, err := client.New(restConfig, client.Options{Scheme: scheme})
	require.NoError(t, err)

	deleteNS := createTestNamespace(t, f)
	in := newTestInstance(t, f)
	require.NoError(t, c.Create(context.Background(), in))
	return in, deleteNS
}

// newTestInstance returns an example instance of an app config.
func newTestInstance(t *testing.T, f testFeatureFlags) *appconfig.AppEnvConfigTemplateV2 {
	namespace := testNamespaceName(t)
	in := &appconfig.AppEnvConfigTemplateV2{
		ObjectMeta: metav1.ObjectMeta{Name: "my-appconfig", Namespace: namespace},
		Spec: appconfig.AppEnvConfigTemplateV2Spec{
			Services: []appconfig.AppEnvConfigTemplateServiceInfo{
				{
					Name:                   "my-service",
					DeploymentApp:          "my-deployment",
					DeploymentVersion:      "v1",
					DeploymentPort:         7000,
					ServicePort:            8000,
					DeploymentPortProtocol: "TCP",
					AllowedClients: []appconfig.AppEnvConfigTemplateRelatedClientInfo{
						{Name: "my-allowed-service-name-0"},
					},
					Ingress: &appconfig.ServiceIngress{
						Host: "my-host",
						Path: "/my-path",
					},
				},
			},
			AllowedEgress: []appconfig.AppEnvConfigTemplateAllowedEgress{
				{Type: "kafka", Hosts: []string{"my.kafka.server"}},
			},
			Auth: &appconfig.AppEnvConfigTemplateAuth{
				JWT: &appconfig.AppEnvConfigTemplateJWT{
					Type: "firebase",
					Params: map[string]string{
						"project": "my-firebase-project",
					},
				},
			},
		},
	}

	if f.vault {
		in.Spec.Auth.GCPAccess = &appconfig.AppEnvConfigTemplateGCPAccess{
			AccessType: "vault",
			VaultInfo: &appconfig.AppEnvConfigTemplateGCPAccessVaultInfo{
				ServiceAccount: "TODO",
				Path:           "TODO",
				Roleset:        "TODO",
			},
		}
	}

	return in
}

// unstructuredShouldExist asserts that a kubernetes object exists with a given
// GroupVersionResource.
func unstructuredShouldExist(t *testing.T, dyn dynamic.Interface, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) {
	c := dyn.Resource(gvr).Namespace(obj.GetNamespace())
	retryTest(t, func() error {
		_, err := c.Get(obj.GetName(), metav1.GetOptions{})
		return err
	})
}

// unstructuredShouldExist asserts that a kubernetes object does NOT exist
// with a given GroupVersionResource.
func unstructuredShouldNotExist(t *testing.T, dyn dynamic.Interface, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) {
	c := dyn.Resource(gvr).Namespace(obj.GetNamespace())
	retryTest(t, func() error {
		_, err := c.Get(obj.GetName(), metav1.GetOptions{})
		return shouldBeNotFound(err)
	})
}

// shouldBeNotFound returns and error if the provided error is not of kind
// NotFound.
func shouldBeNotFound(err error) error {
	if err == nil || !errors.IsNotFound(err) {
		return fmt.Errorf("expected error NotFound, got %v", err)
	}
	return nil
}

// removeServiceFromSpec slices an app config .spec.services[] array.
func removeServiceFromSpec(t *testing.T, c client.Client, in *appconfig.AppEnvConfigTemplateV2, i int) {
	in = in.DeepCopy()
	in.Spec.Services = append(in.Spec.Services[:i], in.Spec.Services[i+1:]...)
	require.NoError(t, c.Update(context.Background(), in))
}

// removeAllowedEgressFromSpec slices an app config .spec.allowedEgress[] array.
func removeAllowedEgressFromSpec(t *testing.T, c client.Client, in *appconfig.AppEnvConfigTemplateV2, i int) {
	in = in.DeepCopy()
	in.Spec.AllowedEgress = append(in.Spec.AllowedEgress[:i], in.Spec.AllowedEgress[i+1:]...)
	require.NoError(t, c.Update(context.Background(), in))
}

// retryTest retries a given function up to 10 times.
func retryTest(t *testing.T, fn func() error) {
	const n = 10
	if err := retry(n, fn); err != nil {
		t.Fatalf("failed after %v attempts: %v", n, err)
	}
}

// retry an operation n times or until error is nil.
func retry(attempts int, fn func() error) error {
	if err := fn(); err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(time.Second / 2)
			return retry(attempts, fn)
		}
		return err
	}
	return nil
}
