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

package builtins

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// getConfigMap finds and returns ConfigMap of a given name in a given namespace
func getConfigMap(ctx context.Context, name, ns string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	cl := localMgr.GetClient()

	if err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, cm); err != nil {
		return cm, fmt.Errorf("finding %s ConfigMap: %s", name, err)
	}

	return cm, nil
}

// injectVolume adds a given volume to the given pod, if not already exists
func injectVolume(pod *corev1.Pod, volume corev1.Volume) {
	log.V(1).Info("injectVolume", "volumeName", volume.Name)

	idx := -1
	for i, v := range pod.Spec.Volumes {
		if v.Name == volume.Name {
			idx = i
			break
		}
	}

	if idx >= 0 {
		log.V(1).Info("injectVolume:Updated", "element.Name", volume.Name)
		pod.Spec.Volumes[idx] = volume
	} else {
		log.V(1).Info("injectVolume:Added", "element.Name", volume.Name)
		pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
	}
}

// injectEnvVar adds a given EnvVar to all containers in a given pod
func injectEnvVar(pod *corev1.Pod, envVar corev1.EnvVar) {
	log.V(1).Info("injectEnvVar", "element.name", envVar.Name)

	find := func(c corev1.Container) int {
		for n, ev := range c.Env {
			if ev.Name == envVar.Name {
				return n
			}
		}
		return -1
	}

	for i, c := range pod.Spec.Containers {
		idx := find(c)

		if idx < 0 {
			log.Info("injectEnvVar:Added", "Container.Name", c.Name, "EnvVar.Name", envVar.Name)
			pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, envVar)
		} else {
			log.Info("injectEnvVar:Updated", "Container.Name", c.Name, "EnvVar.Name", envVar.Name)
			c.Env[idx] = envVar
		}
	}

}

// injectVolumeMount adds a given volumeMount to all containers in a given pod
func injectVolumeMount(pod *corev1.Pod, volumeMount corev1.VolumeMount) {
	log.V(1).Info("injectVolumeMount", "volumeName", volumeMount.Name)

	find := func(c corev1.Container) int {
		for n, vm := range c.VolumeMounts {
			if vm.Name == volumeMount.Name {
				return n
			}
		}
		return -1
	}

	for i, c := range pod.Spec.Containers {
		idx := find(c)

		if idx < 0 {
			log.Info("injectVolumeMount:Added", "Container.Name", c.Name, "VolumeMount.Name", volumeMount.Name)
			pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, volumeMount)
		} else {
			log.Info("injectVolumeMount:Updated", "Container.Name", c.Name, "VolumeMount.Name", volumeMount.Name)
			c.VolumeMounts[idx] = volumeMount
		}
	}

}

// injectVolumeMount adds a given volumeMount to all containers in a given pod
func getVolumeMountsInExistingContainers(pod *corev1.Pod) *corev1.VolumeMount {
	log.V(1).Info("getVolumeMountsInExistingContainers", "volumeName", pod.Name)

	for _, c := range pod.Spec.Containers {
		for _, vm := range c.VolumeMounts {
			if vm.MountPath == "/var/run/secrets/kubernetes.io/serviceaccount" {
				return vm.DeepCopy()
			}
		}
	}

	return nil

}

// injectContainer adds a given container to the given pod, if not already exists
func injectContainer(pod *corev1.Pod, container corev1.Container) {
	log.V(1).Info("injectContainer", "containerName", container.Name)

	idx := -1
	for i, c := range pod.Spec.Containers {
		if c.Name == container.Name {
			idx = i
			break
		}
	}

	if idx >= 0 {
		log.V(1).Info("injectContainer:containerUpdated", "element.Name", container.Name)
		pod.Spec.Containers[idx] = container
	} else {
		log.V(1).Info("injectContainer:containerAdded", "element.Name", container.Name)
		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}
}

// injectInitContainer adds a given init container to the given pod, if not already exists
func injectInitContainer(pod *corev1.Pod, container corev1.Container) {
	log.V(1).Info("injectInitContainer", "containerName", container.Name)

	idx := -1
	for i, c := range pod.Spec.InitContainers {
		if c.Name == container.Name {
			idx = i
			break
		}
	}

	if idx >= 0 {
		log.V(1).Info("injectInitContainer:containerUpdated", "element.Name", container.Name)
		pod.Spec.InitContainers[idx] = container
	} else {
		log.V(1).Info("injectInitContainer:containerAdded", "element.Name", container.Name)
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)
	}
}

// svcAcctJWT looks up the stored JWT secret token for a given service account
func svcAcctJWT(ctx context.Context, name, namespace string) (string, error) {
	log.Info("common:svcAcctJWT")

	var (
		err error

		cl         = localMgr.GetClient()
		secret     = &corev1.Secret{}
		svcAccount = &corev1.ServiceAccount{}
	)

	log.Info("common:svcAcctJWT:secret", "name", name, "namespace", namespace)

	// get service account
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, svcAccount)
	if err != nil {
		log.Error(err, "get ServiceAccount")
		return "", fmt.Errorf("%s serviceAccount not found in %s namespace", name, namespace)
	}
	if len(svcAccount.Secrets) == 0 {
		return "", fmt.Errorf("%s serviceAccount token not found", name)
	}

	log.Info("common:svcAcctJWT:secret:value", "name", name, "namespace", namespace)

	ref := svcAccount.Secrets[0]

	// get service account token secret
	err = cl.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace}, secret)
	if err != nil {
		return "", fmt.Errorf("%s serviceAccount token not found: %s", name, err)
	}

	b := string(secret.Data["token"])
	//b, err := base64.StdEncoding.DecodeString(string(secret.Data["token"]))
	//if err != nil {
	//	return "", err
	//}

	return b, nil
}
