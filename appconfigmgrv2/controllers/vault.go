package controllers

import (
	"context"
	"fmt"

	appconfigmgrv1alpha1 "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
)

// vaultInjectEnabled checks the AppEnvConfigTemplateV2 auth spec for
// existing vaultInfo type and fields with basic validation
func (r *AppEnvConfigTemplateV2Reconciler) vaultInjectEnabled(
	ctx context.Context,
	in *appconfigmgrv1alpha1.AppEnvConfigTemplateV2,
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

func (r *AppEnvConfigTemplateV2Reconciler) reconcileVault(
	ctx context.Context,
	in *appconfigmgrv1alpha1.AppEnvConfigTemplateV2,
) error {
	log.Info("Starting Vault reconcile")
	defer log.Info("Vault reconcile complete")
	return nil
}
