package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	natsv1alpha1 "github.com/zeiss/natz-operator/pkg/client/generated/clientset/internalclientset"
)

type CfgConfig struct {
	User string
}

var CfgCmd = &cobra.Command{
	Use:   "cfg",
	Short: "Manage configs",
	Long:  `Manage configs`,
	RunE:  func(cmd *cobra.Command, args []string) error { return runCfg(cmd.Context()) },
}

func runCfg(_ context.Context) error {
	return nil
}

var CfgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configs",
	Long:  `List configs`,
	RunE:  func(cmd *cobra.Command, args []string) error { return runCfgList(cmd.Context()) },
}

func runCfgList(ctx context.Context) error {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", config.GetKubeConfig())
	if err != nil {
		return err
	}

	clientset, err := natsv1alpha1.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}

	cfgs, err := clientset.Natz().NatsConfigs().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, cfg := range cfgs.Items {
		fmt.Println(cfg)
	}

	return nil
}
