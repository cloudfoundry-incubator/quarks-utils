package cmd

import (
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// Meltdown is the reconciliation backoff duration
func Meltdown(cfg *config.Config) {
	meltdownDuration := viper.GetInt("meltdown-duration")
	cfg.MeltdownDuration = time.Duration(meltdownDuration) * time.Second
	meltdownDurationRequeue := viper.GetInt("meltdown-requeue-after")
	cfg.MeltdownRequeueAfter = time.Duration(meltdownDurationRequeue) * time.Second
}

// MeltdownFlags adds to viper flags
func MeltdownFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.Int("meltdown-duration", 60, "Duration (in seconds) of the meltdown period, in which we postpone further reconciles for the same resource")
	pf.Int("meltdown-requeue-after", 30, "Duration (in seconds) for which we delay the requeuing of the reconcile")
	for _, opt := range []string{"meltdown-duration", "meltdown-requeue-after"} {
		viper.BindPFlag(opt, pf.Lookup(opt))

		argToEnv[opt] = envName(opt)
	}
}
