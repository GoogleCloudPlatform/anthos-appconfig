package controllers

import (
	"context"
	"fmt"
	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
	"github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/webhooks/builtins"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// vaultInjectEnabled checks the AppEnvConfigTemplateV2 auth spec for
// existing vaultInfo type and fields with basic validation
func (r *AppEnvConfigTemplateV2Reconciler) vaultInjectEnabled(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) (bool, error) {
	auth := in.Spec.Auth
	if auth == nil || auth.GCPAccess == nil || auth.GCPAccess.AccessType != "vault" {
		return false, nil
	}

	vaultInfo := auth.GCPAccess.VaultInfo
	if vaultInfo == nil {
		return false, fmt.Errorf("vaultInfo not configured")
	}

	if vaultInfo.ServiceAccount == "" {
		return false, fmt.Errorf("vaultInfo missing serviceAccount key")
	}

	if vaultInfo.Path == "" {
		return false, fmt.Errorf("vaultInfo missing gcpPath key")
	}

	if vaultInfo.Roleset == "" {
		return false, fmt.Errorf("vaultInfo missing gcpRoleset key")
	}

	return true, nil
}

func (r *AppEnvConfigTemplateV2Reconciler) reconcileVaultSupportResources(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {

	secretCopyList := &map[string]*corev1.Secret{
		builtins.VAULT_CA_SECRET_NAME: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      builtins.VAULT_CA_SECRET_NAME,
				Namespace: builtins.TODO_FIND_NAMESPACE,
			},
		},
	}

	if err := r.reconcileSecretsToNamespace(ctx, in, secretCopyList); err != nil {
		return fmt.Errorf("reconciling: %v", err)
	}

	return nil
}

func (r *AppEnvConfigTemplateV2Reconciler) reconcileVault(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {

	if err := r.reconcileVaultSupportResources(ctx, in); err != nil {
		return fmt.Errorf("reconciling: %v", err)
	}

	return nil
}
