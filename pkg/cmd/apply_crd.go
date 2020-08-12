package cmd

import (
	"context"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"

	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
)

// ApplyFn is a function that applies CRDs
type ApplyFn func(ctx context.Context, config *rest.Config) error

// ApplyCRDs calls apply to create the operators CRDs
func ApplyCRDs(ctx context.Context, apply ApplyFn, restConfig *rest.Config) error {
	if viper.GetBool("apply-crd") {
		ctxlog.Info(ctx, "Applying CRDs...")
		err := apply(ctx, restConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyCRDsFlags adds to viper flags
func ApplyCRDsFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.Bool("apply-crd", true, "If true, apply CRDs on start")
	viper.BindPFlag("apply-crd", pf.Lookup("apply-crd"))
	argToEnv["apply-crd"] = "APPLY_CRD"
}
