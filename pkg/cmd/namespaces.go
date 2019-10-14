package cmd

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// Namespaces sets the namespaces for our operators
func Namespaces(cfg *config.Config, log *zap.SugaredLogger) string {
	operatorNamespace := viper.GetString("operator-namespace")
	watchNamespace := viper.GetString("watch-namespace")
	if watchNamespace == "" {
		log.Infof("No watch namespace defined. Falling back to the operator namespace.")
		watchNamespace = operatorNamespace
	}

	cfg.OperatorNamespace = operatorNamespace
	cfg.Namespace = watchNamespace

	return watchNamespace
}

// NamespacesFlags adds to viper flags
func NamespacesFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.StringP("operator-namespace", "n", "default", "The operator namespace")
	pf.StringP("watch-namespace", "", "", "Namespace to watch for BOSH deployments")

	viper.BindPFlag("operator-namespace", pf.Lookup("operator-namespace"))
	viper.BindPFlag("watch-namespace", pf.Lookup("watch-namespace"))

	argToEnv["operator-namespace"] = "OPERATOR_NAMESPACE"
	argToEnv["watch-namespace"] = "WATCH_NAMESPACE"
}
