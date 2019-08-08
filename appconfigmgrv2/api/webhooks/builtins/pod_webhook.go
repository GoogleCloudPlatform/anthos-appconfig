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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	VAULT_CONFIGMAP_NAME = "vault"
	VAULT_CA_SECRET_NAME = "vault-ca"
	TODO_FIND_NAMESPACE  = "appconfigmgrv2-system"
)

var (
	log      = ctrl.Log.WithName("webhooks-builtins-pod")
	localMgr ctrl.Manager
)

func SetupWebHook(mgr ctrl.Manager) {

	// Setup webhooks
	// entryLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	localMgr = mgr

	// entryLog.Info("registering webhooks to the webhook server")
	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: &podAnnotator{}})
	hookServer.Register("/validate-v1-pod", &webhook.Admission{Handler: &podValidator{}})

}

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=upod.appconfigmgr.cft.dev

// podAnnotator annotates Pods
type podAnnotator struct {
	client  client.Client
	decoder *admission.Decoder
}

//func getJSONKey(client client.Client, ns string, secret string) {
//	client.
//}

func kubeSecretFromTemplate(ns string, name string, mapKey string, value string) *corev1.Secret {
	return kubeSecretFromTemplateBytes(ns, name, mapKey, []byte(value))
}

func kubeSecretFromTemplateBytes(ns string, name string, mapKey string, value []byte) *corev1.Secret {
	return &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data: map[string][]byte{
			mapKey: []byte(value),
		},
	}
}

func kubeSecretReponse(ns string, name string) *corev1.Secret {
	return &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
	}
}

func getSecretName(ns string) string {
	log.Info("getSecretName:Start:" + "demo-" + ns + "-key")
	return "demo-" + ns + "-key"
}

func updateSecretsVolume(pod *corev1.Pod, secretName string) {
	log.V(1).Info("updateSecretsVolume", "secretName", secretName)
	found := false
	index := -1
	for i, element := range pod.Spec.Volumes {
		if element.Name == "google-auth-token" {
			log.V(1).Info("updateSecretsVolume:volumeFound", "element.Name", element.Name)
			found = true
			index = i
		}
	}
	element := &corev1.Volume{
		Name: "google-auth-token",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	if !found {
		index = len(pod.Spec.Volumes)
		pod.Spec.Volumes = append(pod.Spec.Volumes, *element)
	} else {
		pod.Spec.Volumes[index] = *element
	}

}

func updateContainers(pod *corev1.Pod, appName string, mountName string, mountPath string, envName string) {
	log.Info("updateContainers",
		"appName", appName,
		"mountName", mountName,
		"mountPath", mountPath,
	)

	for index, element := range pod.Spec.Containers {
		if strings.HasPrefix(appName, element.Name) {
			log.Info("updateContainers:found",
				"appName", element.Name,
				"mountName", mountName,
				"mountPath", mountPath)
			updateContainerMounts(&element, element.Name, mountName, mountPath)
			updateContainerEnv(&element, element.Name, envName, mountPath+"/key.json")
			pod.Spec.Containers[index] = element
			return
		}
	}

	//TODO - Decide how to fail or just warning
	log.Info("updateContainers:containerNotFound", "appName", appName,
		"mountName", mountName,
		"mountPath", mountPath)
}

func updateContainerMounts(container *corev1.Container, containerName string, mountName string, mountPath string) {
	log.V(1).Info("updateContainerMounts",
		"containerName", containerName,
		"mountName", mountName,
		"mountPath", mountPath,
	)
	found := false
	index := -1
	for i, element := range container.VolumeMounts {
		if element.Name == mountName {
			found = true
			index = i
			log.V(1).Info("updateContainerMounts:found",
				"containerName", containerName,
				"mountName", mountName,
				"mountPath", mountPath,
			)
		}
	}
	element := &corev1.VolumeMount{
		Name:      mountName,
		MountPath: mountPath,
	}

	if !found {
		// Append The Mount
		log.V(1).Info("updateContainerMounts:addMount",
			"containerName", containerName,
			"mountName", mountName,
			"mountPath", mountPath,
		)
		index = len(container.VolumeMounts)
		container.VolumeMounts = append(container.VolumeMounts, *element)
	} else {
		container.VolumeMounts[index] = *element
	}

	log.V(1).Info("updateContainerMounts:exit",
		"containerInfo", container,
	)

}

func updateContainerEnv(container *corev1.Container, containerName string, envName string, mountPath string) {
	log.V(1).Info("updateContainerEnv",
		"containerName", containerName,
		"envName", envName,
		"mountPath", mountPath,
	)
	found := false
	index := -1
	for i, element := range container.Env {
		if element.Name == envName {
			found = true
			index = i
			log.V(1).Info("updateContainerEnv:found",
				"containerName", containerName,
				"envName", envName,
				"mountPath", mountPath,
			)
		}
	}
	element := &corev1.EnvVar{
		Name:  envName,
		Value: mountPath,
	}

	if !found {
		// Append The Mount
		log.V(1).Info("updateContainerEnv:addMount",
			"containerName", containerName,
			"envName", envName,
			"mountPath", mountPath,
		)
		index = len(container.Env)
		container.Env = append(container.Env, *element)
	} else {
		container.Env[index] = *element
	}

	log.V(1).Info("updateContainerEnv:exit",
		"containerInfo", container,
	)

}

func (a *podAnnotator) handleGCPSecretIfNeeded(ctx context.Context, pod *corev1.Pod, app *appconfig.AppEnvConfigTemplateV2) error {
	log.Info("podAnnotator:handleGCPSecretIfNeeded")
	switch {
	case app.Spec.Auth == nil, app.Spec.Auth.GCPAccess == nil:
		return nil
	case app.Spec.Auth.GCPAccess.AccessType == "vault":
		return a.handleGCPVault(ctx, pod, app)
	case app.Spec.Auth.GCPAccess.AccessType == "secret":
		return a.handleGCPSecret(ctx, pod, app)
	default:
		log.Error(fmt.Errorf("invalid GCPAccess value"), "\"%s\"", app.Spec.Auth.GCPAccess.AccessType)
		return nil
	}
}

func (a *podAnnotator) handleGCPVault(ctx context.Context, pod *corev1.Pod, app *appconfig.AppEnvConfigTemplateV2) error {
	log.Info("podAnnotator:handleGCPVault")

	var (
		caVolName  = VAULT_CA_SECRET_NAME + "-vol"
		gcpVolName = "google-auth-token"
		vaultInfo  = app.Spec.Auth.GCPAccess.VaultInfo
	)

	log.Info("handleGCPVault:loadConfig")

	// read vaultInfo from AppEnvConfigTemplateV2 spec
	if vaultInfo == nil {
		return fmt.Errorf("vaultInfo not configured")
	}

	if vaultInfo.ServiceAccount == "" {
		return fmt.Errorf("vaultInfo missing serviceAccount field")
	}

	if vaultInfo.Path == "" {
		return fmt.Errorf("vaultInfo missing gcpPath field")
	}

	// construct image name and tag from env
	imageBuild := os.Getenv("CONTROLLER_BUILD")
	if imageBuild == "" {
		imageBuild = "latest"
	}
	imageRegistry := os.Getenv("CONTROLLER_REGISTRY")
	if imageRegistry == "" {
		imageRegistry = "gcr.io/anthos-appconfig"
	}
	image := fmt.Sprintf("%s/vault-api-helper:%s", imageRegistry, imageBuild)

	// get vault configMap, validate
	log.Info("handleGCPVault:loadConfig", "ConfigMap", VAULT_CONFIGMAP_NAME)
	config, err := getConfigMap(ctx, VAULT_CONFIGMAP_NAME, TODO_FIND_NAMESPACE)
	if err != nil {
		return err
	}

	if config.Data["vault-addr"] == "" {
		return fmt.Errorf("ConfigMap missing vault-addr")
	}

	if config.Data["acm-cluster-name"] == "" {
		return fmt.Errorf("ConfigMap missing acm-cluster-name")
	}

	// get provided serviceAccount JWT token
	log.Info("handleGCPVault:loadConfig", "ServiceAccount", vaultInfo.ServiceAccount)
	ksaToken, err := svcAcctJWT(ctx, vaultInfo.ServiceAccount, app.Namespace)
	if err != nil {
		return err
	}
	log.Info("handleGCPVault:loadConfig", "Token", len(ksaToken))

	VAULT_ADDITIONAL_SECRET := "vault-helper-info"
	secretDataMap := &map[string]string{
		"ksa.token": ksaToken,
	}

	createSecret(context.TODO(), VAULT_ADDITIONAL_SECRET, app.Namespace, secretDataMap)
	// copy vault CA cert into app namespace
	VAULT_CA_SECRET_NAME := "vault-ca"
	log.Info("handleGCPVault:applyConfig", "Secret", VAULT_CA_SECRET_NAME)
	if err := cloneSecret(ctx, VAULT_CA_SECRET_NAME, app); err != nil {
		return err
	}

	// add vault CA cert secret to pod volumes
	log.Info("handleGCPVault:applyConfig", "Volume", VAULT_CA_SECRET_NAME)
	injectVolume(pod, corev1.Volume{
		Name: caVolName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: VAULT_CA_SECRET_NAME,
			},
		},
	})




	log.Info("handleGCPVault:applyConfig", "getVolumeMountForToken", gcpVolName)
	serviceAccountVolumeMount := getVolumeMountsInExistingContainers(pod)
	if (serviceAccountVolumeMount == nil) {
		panic(errors.New("Failed to find serviceAccountVolumeMount"))
	}
	log.Info("handleGCPVault:injectInitContainer", "Container", "vault-gcp-auth")

	// inject vault-gcp init container
	injectInitContainer(pod, corev1.Container{
		Name:            "vault-gcp-auth",
		Image:           image,
		ImagePullPolicy: corev1.PullAlways,
		Env: []corev1.EnvVar{
			{
				Name: "MY_POD_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.namespace",
					},
				},
			},

			{
				Name: "MY_POD_SERVICE_ACCOUNT",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "spec.serviceAccountName",
					},
				},
			},
			{
				Name:  "INIT_GCP_KEYPATH",
				Value: fmt.Sprintf("%s/key/%s", vaultInfo.Path, vaultInfo.Roleset),
			},
			{
				Name:  "INIT_K8S_KEYPATH",
				Value: fmt.Sprintf("auth/k8s-%s/role/%s", config.Data["acm-cluster-name"], vaultInfo.Roleset),
			},
			{
				Name:  "VAULT_ADDR",
				Value: config.Data["vault-addr"],
			},
			{
				Name:  "VAULT_CAPATH",
				Value: "/var/run/secrets/vault/ca.pem",
			},
			{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/run/secrets/google/token/key.json",
			},
			{
				Name:  "INIT_K8S_TOKEN_KEYPATH",
				Value: "/var/run/secrets/google/token/ksa.token",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      caVolName,
				MountPath: "/var/run/secrets/vault",
			},
			{
				Name:      gcpVolName,
				MountPath: "/var/run/secrets/google/token",
			},
			*serviceAccountVolumeMount,
		},
	})

	// inject vault-gcp cycle container
	injectContainer(pod, corev1.Container{
		Name:            "vault-gcp-cycle",
		Image:           image,
		ImagePullPolicy: corev1.PullAlways,
		Command:         []string{"watch"},
		Env: []corev1.EnvVar{
			{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/run/secrets/google/token/key.json",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      gcpVolName,
				MountPath: "/var/run/secrets/google/token",
			},
		},
	})

	// add GCP token volume to pod
	log.Info("handleGCPVault:applyConfig", "Volume", gcpVolName)
	injectVolume(pod, corev1.Volume{
		Name: gcpVolName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	})

	// inject volume mount for all pod containers
	log.Info("handleGCPVault:applyConfig", "VolumeMount", gcpVolName)
	injectVolumeMount(pod, corev1.VolumeMount{
		Name:      gcpVolName,
		ReadOnly:  true,
		MountPath: "/var/run/secrets/google/token",
	})

	// inject app credential env var for all pod containers
	log.Info("handleGCPVault:applyConfig", "EnvVar", "GOOGLE_APPLICATION_CREDENTIALS")
	injectEnvVar(pod, corev1.EnvVar{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: "/var/run/secrets/google/token/key.json",
	})

	return nil
}

func (a *podAnnotator) handleGCPSecret(ctx context.Context, pod *corev1.Pod, app *appconfig.AppEnvConfigTemplateV2) error {
	log.Info("podAnnotator:handleGCPSecret")

	secretName := app.Spec.Auth.GCPAccess.SecretInfo.Name
	secretNamespace := TODO_FIND_NAMESPACE
	secret := &corev1.Secret{}

	cl := localMgr.GetClient()
	err := cl.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil {
		log.Error(err, "Get Google Key from Secret to generate token")
		return errors.New("Secret Not Found")
		//	Try Create
		//err = cl.Create(ctx, kubeSecretFromTemplate(req.Namespace, "google-cloud-key"))
		//if err != nil {
		//	log.Error(err, "Secret:Create")
		//	return admission.Errored(http.StatusBadRequest, err)
		//}
	}
	log.Info("HandleUpdate:Secret", "secret", secret.Name)
	token := string(secret.Data["key.json"])

	appSecret := &corev1.Secret{}
	err = cl.Get(ctx, types.NamespacedName{Name: "google-cloud-token", Namespace: app.Namespace}, appSecret)
	if err != nil {
		// avoid using ! in compound statement due to readability
		if k8sapierrors.IsNotFound(err) {
			err = cl.Create(ctx, kubeSecretFromTemplate(app.Namespace, "google-cloud-token", "key.json", token))
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		appSecret.Data["key.json"] = []byte(token)
		err = cl.Update(ctx, appSecret)
		if err != nil {
			return err
		}
	}
	log.Info("HandleUpdate:Volume Mounts", "secret", "google-cloud-token")
	updateSecretsVolume(pod, "google-cloud-token")

	log.Info("HandleUpdate:Containers", "pod.Labels", pod.GetLabels())
	if len(pod.GetLabels()["app"]) > 0 {
		log.Info("HandleUpdate:Containers:app", "pod.Labels.app", pod.GetLabels()["app"])
		updateContainers(pod, pod.GetLabels()["app"], "google-auth-token",
			"/var/run/secrets/google/token", "GOOGLE_APPLICATION_CREDENTIALS")
	}

	return nil

}

func getApplicationName(pod *corev1.Pod) (string, error) {
	if pod.Annotations == nil {
		return "", errors.New("Annotation not found, empty annotations")
	}

	if val, ok := pod.Annotations["appconfigmgr.cft.dev/application"]; ok {
		return val, nil
	}

	return "", errors.New("Annotation not found, empty annotations")
}

// podAnnotator adds an annotation to every incoming pods.
func (a *podAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {

	pod := &corev1.Pod{}

	log.Info("HandleUpdate:Start", req.Name, req.Namespace)

	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	app := &appconfig.AppEnvConfigTemplateV2{}

	applicationName, err := getApplicationName(pod)
	if err != nil {
		log.Error(err, "Application annotation not found")

	}

	log.Info("HandleUpdate:applicationName", "applicationName", applicationName,
		"req.Namespace", req.Namespace, "req.Operation", req.Operation)

	err = localMgr.GetClient().Get(ctx, types.NamespacedName{Name: applicationName, Namespace: req.Namespace}, app)
	if err != nil {
		log.Error(err, "Application Does not Exist - working to see why it is not in scheme, hardcoded app to pubsub")
		//return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == "CREATE" {
		err := a.handleGCPSecretIfNeeded(ctx, pod, app)
		if err != nil {
			log.Error(err, "Application GCP Secret could not be handled see error")
			return admission.Errored(http.StatusBadRequest, err)
		}
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations["example-mutating-admission-webhook"] = "foo"

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// podAnnotator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (a *podAnnotator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// podAnnotator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *podAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// +kubebuilder:webhook:path=/validate-v1-pod,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.appconfigmgr.cft.dev

// podValidator validates Pods
type podValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// podValidator admits a pod iff a specific annotation exists.
func (v *podValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	key := "example-mutating-admission-webhook"
	anno, found := pod.Annotations[key]
	if !found {
		return admission.Denied(fmt.Sprintf("missing annotation %s", key))
	}
	if anno != "foo" {
		return admission.Denied(fmt.Sprintf("annotation %s did not have value %q", key, "foo"))
	}

	return admission.Allowed("")
}

// podValidator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (v *podValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// podValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *podValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
