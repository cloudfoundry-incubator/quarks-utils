package cmd

import (
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"k8s.io/client-go/rest"

	"code.cloudfoundry.org/quarks-utils/pkg/kubeconfig"
)

// KubeConfig uses kubeconfig pkg to return a valid kube config
func KubeConfig(log *zap.SugaredLogger) (*rest.Config, error) {
	restConfig, err := kubeconfig.NewGetter(log).Get(viper.GetString("kubeconfig"))
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't fetch Kubeconfig. Ensure kubeconfig is present to continue.")
	}
	if err := kubeconfig.NewChecker(log).Check(restConfig); err != nil {
		return nil, errors.Wrap(err, "Couldn't check Kubeconfig. Ensure kubeconfig is correct to continue.")
	}
	return restConfig, nil
}

// KubeConfigFlags adds to viper flags
func KubeConfigFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.StringP("kubeconfig", "c", "", "Path to a kubeconfig, not required in-cluster")
	//nolint:errcheck
	viper.BindPFlag("kubeconfig", pf.Lookup("kubeconfig"))
	argToEnv["kubeconfig"] = "KUBECONFIG"
}
