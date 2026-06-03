// Collector operator: controller-runtime in-cluster reconciliation or legacy file-based loop.
package main

import (
	"flag"
	"os"
	"time"

	operatorv1 "github.com/neuralops/platform/internal/operator/api/v1"
	"github.com/neuralops/platform/internal/operator/controller"
	"github.com/neuralops/platform/internal/operator/gateway"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(operatorv1.AddToScheme(scheme))
}

func main() {
	legacySpec := flag.String("spec", os.Getenv("COLLECTOR_RECONCILE_SPEC"), "legacy reconcile JSON (file mode)")
	legacyMode := flag.Bool("legacy", os.Getenv("OPERATOR_LEGACY") == "true", "use file-based reconcile loop instead of controller-runtime")
	metricsAddr := flag.String("metrics-bind-address", ":8082", "metrics bind address")
	probeAddr := flag.String("health-probe-bind-address", ":8083", "health probe bind address")
	requeue := flag.Duration("requeue", 30*time.Second, "reconcile requeue interval")
	flag.Parse()

	if *legacyMode || *legacySpec != "" {
		runLegacy(*legacySpec)
		return
	}

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg, err := rest.InClusterConfig()
	if err != nil {
		setupLog.Error(err, "in-cluster config unavailable; set OPERATOR_LEGACY=true or COLLECTOR_RECONCILE_SPEC for file mode")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{BindAddress: *metricsAddr},
		HealthProbeBindAddress:     *probeAddr,
		LeaderElection:             true,
		LeaderElectionID:           "neuralops-collector-operator",
		LeaderElectionResourceLock: "leases",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controller.CollectorAgentReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		Gateway:      gateway.NewClientFromEnv(),
		RequeueAfter: *requeue,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting controller-runtime collector operator")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
