package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	natzv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/controllers"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var build = fmt.Sprintf("%s (%s) (%s)", version, commit, date)

type flags struct {
	metricsAddr string
	probeAddr   string
}

var f = &flags{}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var rootCmd = &cobra.Command{
	Use:     "account-server",
	Version: build,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context())
	},
}

func init() {
	rootCmd.Flags().StringVar(&f.metricsAddr, "metrics-bind-address", ":8084", "metrics endpoint")
	rootCmd.Flags().StringVar(&f.probeAddr, "health-probe-bind-address", ":8085", "health probe")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(natzv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func run(ctx context.Context) error {
	opts := zap.Options{
		Development: true,
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: f.metricsAddr},
		HealthProbeBindAddress: f.probeAddr,
		LeaderElection:         false,
		BaseContext:            func() context.Context { return ctx },
	})
	if err != nil {
		return err
	}

	nc, err := nats.Connect(os.Getenv("NATS_URL"), nats.UserCredentials(os.Getenv("NATS_CREDS_FILE")))
	if err != nil {
		return err
	}
	defer nc.Drain()
	defer nc.Close()

	ac := controllers.NewNatsAccountServer(mgr, nc)
	err = ac.SetupWithManager(mgr)
	if err != nil {
		return err
	}

	//+kubebuilder:scaffold:builders

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}

	setupLog.Info("starting manager")
	// nolint:contextcheck
	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		setupLog.Error(err, "unable to run operator")
	}
}
