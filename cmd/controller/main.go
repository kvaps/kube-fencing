package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/kvaps/kube-fencing/pkg/controller"
	"github.com/kvaps/kube-fencing/pkg/controller/node"
	"github.com/kvaps/kube-fencing/version"

	//"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func printVersion() {
	klog.Info(fmt.Sprintf("Kube-Fencing Version: %s", version.Version))
	klog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	klog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {

	flag.Parse()
	printVersion()

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	// Get current namespace
	Namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		klog.Errorln("Failed to get watch namespace", err)
		os.Exit(1)
	}
	node.Namespace = Namespace

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorln("Failed to get kubernetes config", err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress:      "0",
		Namespace:               Namespace,
		LeaderElection:          true,
		LeaderElectionID:        "kube-fencing-lock",
		LeaderElectionNamespace: Namespace,
	})
	if err != nil {
		klog.Errorln("Failed to create new manager", err)
		os.Exit(1)
	}

	klog.Infoln("Registering Components.")

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		klog.Errorln("Failed to setup controllers", err)
		os.Exit(1)
	}

	klog.Infoln("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		klog.Errorln("Manager exited non-zero", err)
		os.Exit(1)
	}
}
