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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
	appconfigmgrv1alpha1 "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
)

var log = ctrl.Log.WithName("controller")

// AppEnvConfigTemplateV2Reconciler reconciles a AppEnvConfigTemplateV2 object.
type AppEnvConfigTemplateV2Reconciler struct {
	client.Client

	Dynamic dynamic.Interface
	Log     logr.Logger
	Scheme  *runtime.Scheme

	skipGatekeeper bool
}

// Reconcile takes an instance of an app config and issues create/update/delete requests
// to a number of resources. Behavior is dependant on whether or not istio auto-inject is
// enabled for the namespace.
func (r *AppEnvConfigTemplateV2Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log = r.Log.WithValues("appenvconfigtemplatev2", req.NamespacedName)

	log.Info("Starting reconcile")
	defer log.Info("Reconcile complete")

	// Relies on OPA Gatekeeper.
	if !r.skipGatekeeper {
		/* TODO: Check that app labels are valid via listing instances.
		instanceList := &appconfigmgrv1alpha1.AppEnvConfigTemplateV2List{}
		if err := r.List(ctx, instanceList); err != nil {
			return ctrl.Result{}, err
		}
		*/

		opaNamespaces, err := r.opaNamespaces(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("listing opa namespaces: %v", err)
		}
		if err := r.reconcileOPAContraints(ctx, opaNamespaces); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling opa constraints: %v", err)
		}
	}

	instance := &appconfigmgrv1alpha1.AppEnvConfigTemplateV2{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// If istio is enabled, we will light up certain features and use istio
	// resources rather than native kubernetes resources for other features.
	istioEnabled, err := r.istioAutoInjectEnabled(ctx, instance.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking for istio auto-inject label: %v", err)
	}

	cfg, err := r.getConfig()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting config: %v", err)
	}

	log.Info("Reconciling", "resource", "services")
	if err := r.reconcileServices(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling services: %v", err)
	}

	log.Info("Reconciling", "resource", "ingress")
	if err := r.reconcileIngress(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling ingress: %v", err)
	}

	if istioEnabled {
		log.Info("Reconciling", "resource", "virtualservices")
		if err := r.reconcileIstioVirtualServices(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling istio virtual services: %v", err)
		}

		log.Info("Reconciling", "resource", "policies")
		if err := r.reconcileIstioPolicies(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling istio policies: %v", err)
		}

		log.Info("Reconciling", "resource", "serviceentries")
		if err := r.reconcileIstioServiceEntries(ctx, cfg, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling istio service entries: %v", err)
		}

		log.Info("Reconciling", "resource", "instances")
		if err := r.reconcileIstioInstances(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling istio instances: %v", err)
		}

		log.Info("Reconciling", "resource", "handlers")
		if err := r.reconcileIstioHandlers(ctx, cfg, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling istio handlers: %v", err)
		}

		log.Info("Reconciling", "resource", "rules")
		if err := r.reconcileIstioRules(ctx, cfg, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling istio rules: %v", err)
		}
	} else {
		log.Info("Reconciling", "resource", "networkpolicies")
		if err := r.reconcileNetworkPolicies(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling network policies: %v", err)
		}
	}

	// TODO: Garbage collect istio/non-istio resources on namespace istio injection label update?
	// i.e. NetworkPolicies vs istio Rules

	vaultEnabled, err := r.vaultInjectEnabled(ctx, instance)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking vault gcpaccess config: %v", err)
	}

	if vaultEnabled {
		if err := r.reconcileVault(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling vault: %v", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers the reconciler with a manager.
// The behavior is dependant on whether or not istio is installed.
// This is determined by the presence of istio CRDs.
func (r *AppEnvConfigTemplateV2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr).
		For(&appconfigmgrv1alpha1.AppEnvConfigTemplateV2{}).
		// Watch namespaces for enforcing opa constraints.
		Watches(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{}).
		Owns(&corev1.Service{}).
		Owns(&v1beta1.Ingress{}).
		Owns(&netv1.NetworkPolicy{})

	istioInstalled := true
	for _, t := range istioTypes {
		installed, err := r.resourceInstalled(context.Background(), t.Resource)
		if err != nil {
			return fmt.Errorf("checking if istio crd is installed: %v", err)
		}
		if !installed {
			istioInstalled = false
			break
		}
	}
	log.Info("Determined istio installation status", "installed", istioInstalled)
	if istioInstalled {
		for _, t := range istioTypes {
			c.Owns(gvkObject(t.Kind))
		}
	}

	return c.Complete(r)
}

// resourceInstalled checks if a CRD is installed on the cluster.
func (r *AppEnvConfigTemplateV2Reconciler) resourceInstalled(ctx context.Context, gvr schema.GroupVersionResource) (bool, error) {
	c := r.Dynamic.Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	})
	_, err := c.Get(gvr.Resource+"."+gvr.Group, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// gvkObject returns an empty object with its GroupVersionKind set.
func gvkObject(gvk schema.GroupVersionKind) runtime.Object {
	unst := &unstructured.Unstructured{}
	unst.SetGroupVersionKind(gvk)
	return unst
}

// getConfig currenly returns a hardcoded default configuration.
// TODO: Consider pulling from a kube ConfigMap resources instead.
func (r *AppEnvConfigTemplateV2Reconciler) getConfig() (Config, error) {
	return defaultConfig, nil
}

// istioAutoInjectEnabled checks the current namespace for the
// "istio-injection" = "enabled" label.
func (r *AppEnvConfigTemplateV2Reconciler) istioAutoInjectEnabled(ctx context.Context, namespace string) (bool, error) {
	name := types.NamespacedName{Name: namespace}
	ns := &corev1.Namespace{}
	if err := r.Client.Get(ctx, name, ns); err != nil {
		return false, err
	}
	return ns.Labels["istio-injection"] == "enabled", nil
}

// opaNamespaces returns a list of namespaces to enforce opa constraints on.
func (r *AppEnvConfigTemplateV2Reconciler) opaNamespaces(ctx context.Context) ([]string, error) {
	names := make([]string, 0)

	var list corev1.NamespaceList
	if err := r.Client.List(ctx, &list, client.MatchingLabels(map[string]string{
		"mutating-create-update-pod-appconfig-cft-dev": "enabled",
	})); err != nil {
		return nil, err
	}
	for _, ns := range list.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

// upsertUnstructured creates/updates unstructured objects based on spec
// comparisons.
func (r *AppEnvConfigTemplateV2Reconciler) upsertUnstructured(
	ctx context.Context,
	desired *unstructured.Unstructured,
	gvr schema.GroupVersionResource,
	namespaced bool,
) error {
	var client dynamic.ResourceInterface
	if namespaced {
		client = r.Dynamic.Resource(gvr).Namespace(desired.GetNamespace())
	} else {
		client = r.Dynamic.Resource(gvr)
	}

	found, err := client.Get(desired.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating", "resource", gvr.Resource, "name", desired.GetName())
			if _, err := client.Create(desired, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("creating: %v", err)
			}
			return nil
		}

		return fmt.Errorf("getting: %v", err)
	}

	if !reflect.DeepEqual(desired.Object["spec"], found.Object["spec"]) {
		found.Object["spec"] = desired.Object["spec"]
		log.Info("Updating", "resource", gvr.Resource, "name", desired.GetName())
		if _, err := client.Update(found, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("updating: %v", err)
		}
		return nil
	}

	return nil
}

// garbageCollect removes objects that are not contained within the
// provided map.
func (r *AppEnvConfigTemplateV2Reconciler) garbageCollect(
	t *appconfig.AppEnvConfigTemplateV2,
	names map[types.NamespacedName]bool,
	gvr schema.GroupVersionResource,
) error {
	list, err := r.Dynamic.Resource(gvr).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing: %v", err)
	}

	for _, item := range list.Items {
		if !metav1.IsControlledBy(&item, t) {
			continue
		}

		nn := types.NamespacedName{Name: item.GetName(), Namespace: item.GetNamespace()}
		if !names[nn] {
			log.Info("Deleting", "resource", gvr.Resource, "name", nn.Name)
			if err := r.Dynamic.Resource(gvr).
				Namespace(nn.Namespace).
				Delete(nn.Name, &metav1.DeleteOptions{}); err != nil {
				return fmt.Errorf("deleting: %v", err)
			}
		}
	}

	return nil
}
