package builtins

import (
	"context"
	"encoding/base64"
	"fmt"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func getConfigMap(ctx context.Context, name, ns string, cm *corev1.ConfigMap) error {
	cl := localMgr.GetClient()

	if err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, cm); err != nil {
		return fmt.Errorf("finding %s ConfigMap: %s", name, err)
	}

	return nil
}

// injectVolume adds a given volume to the given pod, if not already exists
func injectVolume(pod *corev1.Pod, volume *corev1.Volume) {
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
		pod.Spec.Volumes[idx] = *volume
	} else {
		log.V(1).Info("injectVolume:Added", "element.Name", volume.Name)
		pod.Spec.Volumes = append(pod.Spec.Volumes, *volume)
	}
}

// injectVolumeMount adds a given volumeMount to all containers in a given pod
func injectVolumeMount(pod *corev1.Pod, volumeMount *corev1.VolumeMount) {
	log.V(1).Info("injectVolumeMount", "volumeName", volumeMount.Name)

	mountIdx := func(c corev1.Container, name string) int {
		for n, vm := range c.VolumeMounts {
			if vm.Name == name {
				return n
			}
		}
		return -1
	}

	for _, c := range pod.Spec.Containers {
		idx := mountIdx(c, volumeMount.Name)

		if idx < 0 {
			log.V(1).Info("injectVolumeMount:Added", "Container.Name", c.Name, "VolumeMount.Name", volumeMount.Name)
			c.VolumeMounts = append(c.VolumeMounts, *volumeMount)
		} else {
			log.V(1).Info("injectVolumeMount:Updated", "Container.Name", c.Name, "VolumeMount.Name", volumeMount.Name)
			c.VolumeMounts[idx] = *volumeMount
		}
	}

}

// injectContainer adds a given container to the given pod, if not already exists
func injectContainer(pod *corev1.Pod, container *corev1.Container) {
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
		pod.Spec.Containers[idx] = *container
	} else {
		log.V(1).Info("injectContainer:containerAdded", "element.Name", container.Name)
		pod.Spec.Containers = append(pod.Spec.Containers, *container)
	}
}

// injectInitContainer adds a given init container to the given pod, if not already exists
func injectInitContainer(pod *corev1.Pod, container *corev1.Container) {
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
		pod.Spec.InitContainers[idx] = *container
	} else {
		log.V(1).Info("injectInitContainer:containerAdded", "element.Name", container.Name)
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, *container)
	}
}

// copySecret ensures a duplicate of a secret existing in appconfigmgr
// namespace exists in the app namespace
func copySecret(ctx context.Context, name string, app *appconfig.AppEnvConfigTemplateV2) error {
	var (
		err       error
		create    bool
		secret    *corev1.Secret
		appSecret *corev1.Secret

		cl = localMgr.GetClient()
	)

	// get source secret
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: TODO_FIND_NAMESPACE}, secret)
	if err != nil {
		log.Error(err, "get appconfig secret")
		return fmt.Errorf("%s secret not found", name)
	}

	// copy into app namespace if needed
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, appSecret)
	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			return err
		}

		// new secret
		create = true
		appSecret = &corev1.Secret{
			Type: secret.Type,
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: app.Namespace,
			},
		}
	}

	// duplicate secret body
	appSecret.Data = make(map[string][]byte)
	appSecret.StringData = make(map[string]string)
	for k, v := range secret.Data {
		appSecret.Data[k] = v
	}
	for k, v := range secret.StringData {
		appSecret.StringData[k] = v
	}

	if create {
		return cl.Create(ctx, appSecret)
	}
	return cl.Update(ctx, appSecret)
}

// svcAcctJWT looks up the stored JWT secret token for a given service account
func svcAcctJWT(ctx context.Context, name, namespace string) (string, error) {
	var (
		err        error
		secret     *corev1.Secret
		svcAccount *corev1.ServiceAccount

		cl = localMgr.GetClient()
	)

	// get service account
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, svcAccount)
	if err != nil {
		log.Error(err, "get ServiceAccount")
		return "", fmt.Errorf("%s serviceAccount not found in %s namespace", name, namespace)
	}
	if len(svcAccount.Secrets) == 0 {
		return "", fmt.Errorf("%s serviceAccount token not found", name)
	}

	ref := svcAccount.Secrets[0]

	// get service account token secret
	err = cl.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, secret)
	if err != nil {
		return "", fmt.Errorf("%s serviceAccount token not found: %s", name, err)
	}

	b, err := base64.StdEncoding.DecodeString(string(secret.Data["token"]))
	if err != nil {
		return "", err
	}

	return string(b), nil
}
