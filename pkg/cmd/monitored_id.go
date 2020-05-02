package cmd

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// MonitoredID sets the moitored id from viper
func MonitoredID(cfg *config.Config) {
	id := viper.GetString("monitored-id")
	if id == "" {
		id = "default"
	}
	cfg.MonitoredID = id
}

// MonitoredIDFlags adds to viper flags
func MonitoredIDFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.String("monitored-id", "default", "only monitor namespaces with this id in their namespace label")
	viper.BindPFlag("monitored-id", pf.Lookup("monitored-id"))
	argToEnv["monitored-id"] = "MONITORED_ID"
}
