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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileSecrets reconciles kubernetes Secrets resources to stand in front of
// deployments which are managed outside the scope of this controller.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileSecretsToNamespace(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
	secretsCopyList *map[string]*corev1.Secret,
) error {

	names := make(map[types.NamespacedName]bool)

	for k, s := range *secretsCopyList {
		toSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      s.Name,
				Namespace: in.Namespace,
			},
		}
		if err := controllerutil.SetControllerReference(in, toSecret, r.Scheme); err != nil {
			return fmt.Errorf("setting controller reference for secret[%v]: %v", k, err)
		}

		log.Info("Reconciling", "resource", "secrets", "key", k, "name", s.Name, "namespace", s.Namespace)
		if err := r.reconcileSecret(ctx, s, toSecret); err != nil {
			return fmt.Errorf("reconciling secret[%v]: %v", k, err)
		}

		names[types.NamespacedName{Name: s.Name, Namespace: s.Namespace}] = true
	}

	if err := r.garbageCollectSecrets(ctx, in, names); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) reconcileSecret(
	ctx context.Context,
	s *corev1.Secret,
	toSecret *corev1.Secret,
) error {
	foundOriginal := &corev1.Secret{}

	log.Info("Check", "resource", "secrets", "namespace", s.Namespace, "name", s.Name)

	err := r.Get(ctx, types.NamespacedName{Name: s.Name, Namespace: s.Namespace}, foundOriginal)
	if err != nil {
		log.Error(err, "resource", "namespace", s.Namespace, "name", s.Name)
		return err
	}

	found := &corev1.Secret{}

	err = r.Get(ctx, types.NamespacedName{Name: toSecret.Name, Namespace: toSecret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Copy", "resource", "secrets", "namespace", toSecret.Namespace, "name", toSecret.Name)
		found = foundOriginal.DeepCopy()
		found.Namespace = toSecret.Namespace
		found.SetResourceVersion("")
		err = r.Create(ctx, found)
		return err
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(foundOriginal, found) {
		// ClusterIP is assigned after creation when it is not originally set
		// so we will preserve the value.
		found = foundOriginal.DeepCopy()
		found.Namespace = s.Namespace
		found.SetResourceVersion("")
		log.Info("Updating", "resource", "services", "namespace", toSecret.Namespace, "name", toSecret.Name)
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) garbageCollectSecrets(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
	names map[types.NamespacedName]bool,
) error {
	//var list corev1.SecretList
	//if err := r.List(ctx, &list, func(opt *client.ListOptions) {}); err != nil {
	//	return fmt.Errorf("listing: %v", err)
	//}
	//
	//for _, s := range list.Items {
	//	if !metav1.IsControlledBy(&s, in) {
	//		continue
	//	}
	//	if !names[types.NamespacedName{Name: s.Name, Namespace: s.Namespace}] {
	//		log.Info("Deleting", "resource", "secret", "namespace", s.Namespace, "name", s.Name)
	//		if err := r.Delete(ctx, &s); err != nil {
	//			return fmt.Errorf("deleting: %v", err)
	//		}
	//	}
	//}

	return nil
}

//
//func secrets(t *appconfig.AppEnvConfigTemplateV2) []*corev1.Secret {
//	var list []*corev1.Secret
//
//	for i := range t.a {
//		s := &corev1.Secret{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      secretName(t, i),
//				Namespace: t.Namespace,
//			},
//			Spec: corev1.ServiceSpec{
//				Selector: map[string]string{
//					"app": t.Spec.Services[i].DeploymentApp,
//				},
//				Ports: []corev1.ServicePort{
//					{
//						// NOTE: Istio requires prefixed port names such as `http-___`.
//						Name:     "http-default",
//						Protocol: t.Spec.Services[i].DeploymentPortProtocol,
//						Port:     t.Spec.Services[i].ServicePort,
//						TargetPort: intstr.IntOrString{
//							IntVal: t.Spec.Services[i].DeploymentPort,
//						},
//					},
//				},
//			},
//		}
//		list = append(list, s)
//	}
//
//	return list
//}
//
//func secretName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
//	return fmt.Sprintf("%v-%v", t.Name, t.Spec.Services[i].Name)
//}
