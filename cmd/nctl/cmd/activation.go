package cmd

import (
	"context"
	"fmt"

	natsv1alpha1 "github.com/zeiss/natz-operator/pkg/client/generated/clientset/internalclientset"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type ActivationConfig struct {
	Activation string
}

var ActivationCmd = &cobra.Command{
	Use:   "activation",
	Short: "Manage activation",
	Long:  `Manage activation`,
	RunE:  func(cmd *cobra.Command, args []string) error { return nil },
}

var GetActivationCmd = &cobra.Command{
	Use:   "get",
	Short: "Get activation",
	Long:  `Get activation`,
	RunE:  func(cmd *cobra.Command, args []string) error { return runGetActivation(cmd.Context()) },
}

func runGetActivation(ctx context.Context) error {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", config.GetKubeConfig())
	if err != nil {
		return err
	}

	clientset, err := natsv1alpha1.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}

	token, err := clientset.Natz().NatsActivations(config.Namespace).Get(ctx, config.Activation.Activation, metav1.GetOptions{})
	if err != nil {
		return err
	}

	fmt.Println(token.Status.JWT)

	return nil
}
