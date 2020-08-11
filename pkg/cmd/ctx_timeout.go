package cmd

import (
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// CtxTimeOut sets the context timeout from viper
func CtxTimeOut(cfg *config.Config) {
	ctxTimeOut := viper.GetInt("ctx-timeout")
	cfg.CtxTimeOut = time.Duration(ctxTimeOut) * time.Second
}

// CtxTimeOutFlags adds to viper flags
func CtxTimeOutFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.Int("ctx-timeout", 300, "context timeout for each k8s API request in seconds")
	//nolint:errcheck
	viper.BindPFlag("ctx-timeout", pf.Lookup("ctx-timeout"))
	argToEnv["ctx-timeout"] = "CTX_TIMEOUT"
}
