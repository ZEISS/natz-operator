package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/pkg/conv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type CredsConfig struct {
	User string
}

var CredsCmd = &cobra.Command{
	Use:   "creds",
	Short: "Manage credentials",
	Long:  `Manage credentials`,
	RunE:  func(cmd *cobra.Command, args []string) error { return runCreds(cmd.Context()) },
}

func runCreds(ctx context.Context) error {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", config.GetKubeConfig())
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}

	secret, err := clientset.CoreV1().Secrets("default").Get(ctx, config.Creds.User, metav1.GetOptions{})
	if err != nil {
		return err
	}

	fmt.Println(conv.String(secret.Data[v1alpha1.SecretUserCredsKey]))

	return nil
}
