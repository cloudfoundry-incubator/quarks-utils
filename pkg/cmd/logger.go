package cmd

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// LogLevel returns the log-level argument
func LogLevel() string {
	return viper.GetString("log-level")
}

// LoggerFlags adds to viper flags
func LoggerFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.StringP("log-level", "l", "debug", "Only print log messages from this level onward (trace,debug,info,warn)")
	//nolint:errcheck
	viper.BindPFlag("log-level", pf.Lookup("log-level"))
	argToEnv["log-level"] = "LOG_LEVEL"
}
