package builtins

import (
	"bytes"
	"context"
	"fmt"
	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	for _, c := range pod.Spec.Containers {
		idx := find(c)

		if idx < 0 {
			log.Info("injectEnvVar:Added", "Container.Name", c.Name, "EnvVar.Name", envVar.Name)
			c.Env = append(c.Env, envVar)
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

	for _, c := range pod.Spec.Containers {
		idx := find(c)

		if idx < 0 {
			log.Info("injectVolumeMount:Added", "Container.Name", c.Name, "VolumeMount.Name", volumeMount.Name)
			c.VolumeMounts = append(c.VolumeMounts, volumeMount)
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

// cloneSecret ensures a duplicate of a secret existing in appconfigmgr
// namespace exists in the app namespace
func cloneSecret(ctx context.Context, name string, app *appconfig.AppEnvConfigTemplateV2) error {
	var (
		err    error
		create bool
		update bool

		cl        = localMgr.GetClient()
		secret    = &corev1.Secret{}
		appSecret = &corev1.Secret{}
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

	// duplicate secret type
	if appSecret.Type != secret.Type {
		update = true
		appSecret.Type = secret.Type
	}

	// duplicate secret body
	appSecret.Data = make(map[string][]byte)
	appSecret.StringData = make(map[string]string)
	for k, v := range secret.Data {
		if !bytes.Equal(appSecret.Data[k], v) {
			update = true
		}
		appSecret.Data[k] = v
	}
	for k, v := range secret.StringData {
		if appSecret.StringData[k] != v {
			update = true
		}
		appSecret.StringData[k] = v
	}

	if create {
		err = cl.Create(ctx, appSecret)
		if err == nil {
			log.V(1).Info("cloneSecret:secretCreated", "element.Name", name)
		}
		return err
	}

	if update {
		err = cl.Update(ctx, appSecret)
		if err == nil {
			log.V(1).Info("cloneSecret:secretUpdated", "element.Name", name)
		}
		return err
	}

	return nil
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

	return string(b), nil
}

// create secret creates a simple secret
func createSecret(ctx context.Context, name string, namespace string, secretData *map[string]string) error {
	var (
		err          error
		cl           = localMgr.GetClient()
		secret       = &corev1.Secret{}
		newSecretMap map[string][]byte
	)

	// get source secret
	err = cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, secret)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		return fmt.Errorf("%s/%s secret not found", name, namespace)
	}

	for k, v := range *secretData {
		newSecretMap[k] = []byte(v)
	}

	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			return err
		}

		// new secret
		secret = &corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: newSecretMap,
		}
		err = cl.Create(ctx, secret)
		if err == nil {
			log.V(1).Info("cloneSecret:secretCreated", "element.Name", name)
		}
		return err

	} else {

		secret.Data = newSecretMap
		err = cl.Update(ctx, secret)
		if err == nil {
			log.V(1).Info("cloneSecret:secretUpdated", "element.Name", name)
		}
		return err

	}

}
