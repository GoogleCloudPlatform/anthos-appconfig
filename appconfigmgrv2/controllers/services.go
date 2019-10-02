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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileServices reconciles kubernetes Service resources to stand in front of
// deployments which are managed outside the scope of this controller.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileServices(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	names := make(map[types.NamespacedName]bool)

	for i, s := range services(in) {
		if err := controllerutil.SetControllerReference(in, s, r.Scheme); err != nil {
			return fmt.Errorf("setting controller reference for service[%v]: %v", i, err)
		}

		log.Info("Reconciling", "resource", "services", "index", i, "name", s.Name, "namespace", s.Namespace)
		if err := r.reconcileService(ctx, s); err != nil {
			return fmt.Errorf("reconciling service[%v]: %v", i, err)
		}

		names[types.NamespacedName{Name: s.Name, Namespace: s.Namespace}] = true
	}

	if err := r.garbageCollectServices(ctx, in, names); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) reconcileService(
	ctx context.Context,
	s *corev1.Service,
) error {
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: s.Name, Namespace: s.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating", "resource", "services", "namespace", s.Namespace, "name", s.Name)
		err = r.Create(ctx, s)
		return err
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(s.Spec, found.Spec) {
		// ClusterIP is assigned after creation when it is not originally set
		// so we will preserve the value.
		s.Spec.ClusterIP = found.Spec.ClusterIP
		if s.Spec.Type != corev1.ServiceTypeClusterIP {
			for i := range s.Spec.Ports {
				s.Spec.Ports[i].NodePort = found.Spec.Ports[i].NodePort
			}
		}
		found.Spec = s.Spec
		log.Info("Updating", "resource", "services", "namespace", s.Namespace, "name", s.Name)
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) garbageCollectServices(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
	names map[types.NamespacedName]bool,
) error {
	var list corev1.ServiceList
	if err := r.List(ctx, &list, func(opt *client.ListOptions) {}); err != nil {
		return fmt.Errorf("listing: %v", err)
	}

	for _, s := range list.Items {
		if !metav1.IsControlledBy(&s, in) {
			continue
		}
		if !names[types.NamespacedName{Name: s.Name, Namespace: s.Namespace}] {
			log.Info("Deleting", "resource", "services", "namespace", s.Namespace, "name", s.Name)
			if err := r.Delete(ctx, &s); err != nil {
				return fmt.Errorf("deleting: %v", err)
			}
		}
	}

	return nil
}

// services returns a list of kube services that should exist.
// The number of services corresponds 1:1 with the number of .spec.services[]
// that are specified. Service selectors are based on `app` labels.
func services(t *appconfig.AppEnvConfigTemplateV2) []*corev1.Service {
	var list []*corev1.Service

	for i := range t.Spec.Services {
		typ := corev1.ServiceTypeClusterIP
		if t.Spec.Services[i].Ingress != nil {
			typ = corev1.ServiceTypeNodePort
		}
		s := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName(t, i),
				Namespace: t.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Type: typ,
				Selector: map[string]string{
					"app": t.Spec.Services[i].DeploymentApp,
				},
				Ports: []corev1.ServicePort{
					{
						// NOTE: Istio requires prefixed port names such as `http-___`.
						Name:     "http-default",
						Protocol: t.Spec.Services[i].DeploymentPortProtocol,
						Port:     t.Spec.Services[i].ServicePort,
						TargetPort: intstr.IntOrString{
							IntVal: t.Spec.Services[i].DeploymentPort,
						},
					},
				},
			},
		}
		list = append(list, s)
	}

	return list
}

// serviceName returns the name of a given service where i is the index
// .spec.services[i] because the name is derived from the app config name
// and the service name.
func serviceName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-%v", t.Name, t.Spec.Services[i].Name)
}
