package cmd

import (
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// Namespaces sets the namespaces for our operators
func Namespaces(cfg *config.Config, log *zap.SugaredLogger, name string) string {
	operatorNamespace := viper.GetString(name)
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
func NamespacesFlags(pf *flag.FlagSet, argToEnv map[string]string, name string) {
	pf.StringP(name, "n", "default", "The operator namespace")
	pf.StringP("watch-namespace", "", "", "Namespace to watch for BOSH deployments")

	viper.BindPFlag(name, pf.Lookup(name))
	viper.BindPFlag("watch-namespace", pf.Lookup("watch-namespace"))

	argToEnv[name] = envName(name)
	argToEnv["watch-namespace"] = "WATCH_NAMESPACE"
}

func envName(name string) string {
	n := strings.ToUpper(name)
	return strings.Replace(n, "-", "_", -1)
}
