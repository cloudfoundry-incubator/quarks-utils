package cmd

import (
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// OperatorNamespace is the namespace of the service, which points to the webhook server
func OperatorNamespace(cfg *config.Config, log *zap.SugaredLogger, name string) string {
	operatorNamespace := viper.GetString(name)

	cfg.OperatorNamespace = operatorNamespace

	return operatorNamespace
}

// OperatorNamespaceFlags adds to viper flags
func OperatorNamespaceFlags(pf *flag.FlagSet, argToEnv map[string]string, name string) {
	pf.StringP(name, "n", "default", "The operator namespace, for the webhook service")

	//nolint:errcheck
	viper.BindPFlag(name, pf.Lookup(name))

	argToEnv[name] = envName(name)
}

func envName(name string) string {
	n := strings.ToUpper(name)
	return strings.Replace(n, "-", "_", -1)
}
