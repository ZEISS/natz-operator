package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/spf13/cobra"
	natzv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/controllers"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var build = fmt.Sprintf("%s (%s) (%s)", version, commit, date)

type flags struct {
	enableLeaderElection bool
	metricsAddr          string
	probeAddr            string
	secureMetrics        bool
	enableHTTP2          bool
}

var f = &flags{}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var rootCmd = &cobra.Command{
	Use:     "operator",
	Version: build,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context())
	},
}

func init() {
	rootCmd.Flags().BoolVar(&f.enableLeaderElection, "leader-elect", f.enableLeaderElection, "only one controller")
	rootCmd.Flags().StringVar(&f.metricsAddr, "metrics-bind-address", ":8080", "metrics endpoint")
	rootCmd.Flags().StringVar(&f.probeAddr, "health-probe-bind-address", ":8081", "health probe")
	rootCmd.Flags().BoolVar(&f.secureMetrics, "secure-metrics", f.secureMetrics, "serve metrics over https")
	rootCmd.Flags().BoolVar(&f.enableHTTP2, "enable-http2", f.enableHTTP2, "enable http/2")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(natzv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func run(ctx context.Context) error {
	opts := zap.Options{
		Development: true,
	}

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !f.enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   f.metricsAddr,
			SecureServing: f.secureMetrics,
			TLSOpts:       tlsOpts,
		},
		HealthProbeBindAddress: f.probeAddr,
		LeaderElection:         f.enableLeaderElection,
		LeaderElectionID:       "432c802.zeiss.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
		BaseContext: func() context.Context { return ctx },
	})
	if err != nil {
		return err
	}

	err = setupControllers(mgr)
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

func setupControllers(mgr ctrl.Manager) error {
	err := controllers.NewNatsOperatorReconciler(mgr).SetupWithManager(mgr)
	if err != nil {
		return err
	}

	err = controllers.NewNatsAccountReconciler(mgr).SetupWithManager(mgr)
	if err != nil {
		return err
	}

	err = controllers.NewNatsUserReconciler(mgr).SetupWithManager(mgr)
	if err != nil {
		return err
	}

	err = controllers.NewNatsGatewayReconciler(mgr).SetupWithManager(mgr)
	if err != nil {
		return err
	}

	err = controllers.NewNatsKeyReconciler(mgr).SetupWithManager(mgr)
	if err != nil {
		return err
	}

	err = controllers.NewNatsConfigReconciler(mgr).SetupWithManager(mgr)
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
