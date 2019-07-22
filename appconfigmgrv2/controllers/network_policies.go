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

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileNetworkPolicies reconciles kubernetes NetworkPolicy resources to enforce
// transport-level allowedClients support in the absence of istio.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileNetworkPolicies(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	names := make(map[types.NamespacedName]bool)

	nps, err := networkPolicies(in)
	if err != nil {
		return fmt.Errorf("building policies: %v", err)
	}
	for _, np := range nps {
		if err := controllerutil.SetControllerReference(in, np, r.Scheme); err != nil {
			return fmt.Errorf("setting controller reference: %v", err)
		}

		if err := r.reconcileNetworkPolicy(ctx, np); err != nil {
			return fmt.Errorf("reconciling: %v", err)
		}

		names[types.NamespacedName{Name: np.Name, Namespace: np.Namespace}] = true
	}

	if err := r.garbageCollectNetworkPolicies(ctx, in, names); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) reconcileNetworkPolicy(
	ctx context.Context,
	desired *netv1.NetworkPolicy,
) error {
	found := &netv1.NetworkPolicy{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desired.Name,
		Namespace: desired.Namespace,
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating", "resource", "networkpolicies", "namespace", found.Namespace, "name", found.Name)
			if err := r.Create(ctx, desired); err != nil {
				return fmt.Errorf("creating: %v", err)
			}
			return nil
		} else {
			return fmt.Errorf("getting: %v", err)
		}
	}

	if !reflect.DeepEqual(desired.Spec, found.Spec) {
		found.Spec = desired.Spec
		log.Info("Updating", "resource", "networkpolicies", "namespace", found.Namespace, "name", found.Name)
		if err := r.Update(ctx, found); err != nil {
			return fmt.Errorf("updating: %v", err)
		}
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) garbageCollectNetworkPolicies(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
	names map[types.NamespacedName]bool,
) error {
	var list netv1.NetworkPolicyList
	if err := r.List(ctx, &list, func(opt *client.ListOptions) {}); err != nil {
		return fmt.Errorf("listing: %v", err)
	}

	for _, np := range list.Items {
		if !metav1.IsControlledBy(&np, in) {
			continue
		}
		if !names[types.NamespacedName{Name: np.Name, Namespace: np.Namespace}] {
			log.Info("Deleting", "resource", "networkpolicies", "namespace", np.Namespace, "name", np.Name)
			if err := r.Delete(ctx, &np); err != nil {
				return fmt.Errorf("deleting: %v", err)
			}
		}
	}

	return nil
}

func networkPolicies(in *appconfig.AppEnvConfigTemplateV2) ([]*netv1.NetworkPolicy, error) {
	var ps []*netv1.NetworkPolicy

	for i := range in.Spec.Services {
		if len(in.Spec.Services[i].AllowedClients) == 0 {
			continue
		}

		clients := make([]netv1.NetworkPolicyPeer, 0)
		for _, c := range in.Spec.Services[i].AllowedClients {
			ns, app, err := parseAllowedClient(c.Name, in.Namespace)
			if err != nil {
				return nil, fmt.Errorf("parsing allowed client: %v", err)
			}

			// TODO: What to do with namespace? You can only select namespaces by labels,
			// not name.
			_ = ns

			clients = append(clients, netv1.NetworkPolicyPeer{
				NamespaceSelector: nil,
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": app},
				},
			})
		}

		ps = append(ps, &netv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      networkPolicyName(in, i),
				Namespace: in.Namespace,
			},
			Spec: netv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": in.Spec.Services[i].DeploymentApp,
					},
				},
				Ingress: []netv1.NetworkPolicyIngressRule{
					{
						From: clients,
					},
				},
			},
		})
	}
	return ps, nil
}

func networkPolicyName(in *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-service--%v", in.Name, in.Spec.Services[i].Name)
}
