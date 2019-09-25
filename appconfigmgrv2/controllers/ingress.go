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
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIngress reconciles kubernetes Ingress resources.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileIngress(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	ing := ingress(in)
	if ing == nil {
		if err := r.removeIngress(ctx, in); err != nil {
			return fmt.Errorf("garbage collecting: %v", err)
		}
		return nil
	}

	if err := controllerutil.SetControllerReference(in, ing, r.Scheme); err != nil {
		return fmt.Errorf("setting controller reference for ingress: %v", err)
	}

	log.Info("Reconciling", "resource", "ingress", "name", ing.Name, "namespace", ing.Namespace)
	found := &v1beta1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: ing.Name, Namespace: ing.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating", "resource", "ingress", "namespace", ing.Namespace, "name", ing.Name)
		err = r.Create(ctx, ing)
		return err
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(ing.Spec, found.Spec) {
		log.Info("Updating", "resource", "ingress", "namespace", ing.Namespace, "name", ing.Name)
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}

	return nil
}

// removeIngress if it exists.
func (r *AppEnvConfigTemplateV2Reconciler) removeIngress(
	ctx context.Context,
	t *appconfig.AppEnvConfigTemplateV2,
) error {
	meta := ingressMeta(t)
	err := r.Get(ctx, types.NamespacedName{Name: meta.Name, Namespace: meta.Namespace}, &v1beta1.Ingress{})
	if err != nil && errors.IsNotFound(err) {
		// Should not exist, we are good.
		return nil
	} else if err != nil {
		// Error issuing GET.
		return err
	}
	// Exists but should not. Garbage collect.
	log.Info("Deleting", "resource", "ingress", "namespace", meta.Namespace, "name", meta.Name)
	ing := &v1beta1.Ingress{ObjectMeta: meta}
	if err := r.Delete(ctx, ing); err != nil {
		return fmt.Errorf("deleting: %v", err)
	}

	return nil
}

// ingress builds an ingress resource with rules derived from
// from .spec.services[].ingress fields. TLS info is dervied from the
// .spec.ingress.tls field.
// NOTE: Returns nil if no ingress should be created.
func ingress(t *appconfig.AppEnvConfigTemplateV2) *v1beta1.Ingress {
	var rules []v1beta1.IngressRule
	for i, s := range t.Spec.Services {
		if s.Ingress == nil {
			continue
		}

		r := v1beta1.IngressRule{
			Host: s.Ingress.Host,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{{
						Path: s.Ingress.Path,
						Backend: v1beta1.IngressBackend{
							ServiceName: serviceName(t, i),
							ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: s.ServicePort},
						},
					}},
				},
			},
		}
		rules = append(rules, r)
	}

	// No ingress resource should be created.
	if len(rules) == 0 {
		return nil
	}

	tls := []v1beta1.IngressTLS{}
	for _, name := range t.Spec.Ingress.TLS.CertSecrets {
		tls = append(tls, v1beta1.IngressTLS{SecretName: name})
	}

	ing := &v1beta1.Ingress{
		ObjectMeta: ingressMeta(t),
		Spec: v1beta1.IngressSpec{
			TLS:   tls,
			Rules: rules,
		},
	}

	return ing
}

// ingressMeta returns a populated name and namespace.
func ingressMeta(t *appconfig.AppEnvConfigTemplateV2) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%v", t.Name),
		Namespace: t.Namespace,
	}
}
