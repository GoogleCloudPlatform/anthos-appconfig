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

package main

import (
	"flag"
	"os"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
	"github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/webhooks"
	"github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/controllers"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	corev1.AddToScheme(scheme)
	netv1.AddToScheme(scheme)
	////metav1.AddToScheme(scheme)
	//scheme.AllKnownTypes()
	appconfig.AddToScheme(scheme)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	dyn, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create new dynamic client")
		os.Exit(1)
	}

	err = (&controllers.AppEnvConfigTemplateV2Reconciler{
		Client:  mgr.GetClient(),
		Dynamic: dyn,
		Log:     ctrl.Log.WithName("controllers").WithName("AppEnvConfigTemplateV2"),
		Scheme:  mgr.GetScheme(),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AppEnvConfigTemplateV2")
		os.Exit(1)
	}

	webhooks.SetupWebHooks(mgr)

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
