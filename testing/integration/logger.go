package environment

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"

	crlog "sigs.k8s.io/controller-runtime/pkg/log"

	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	helper "code.cloudfoundry.org/quarks-utils/testing/testhelper"
)

// SetupLoggerContext sets up the logger and puts it into a new context
func (e *Environment) SetupLoggerContext(prefix string) context.Context {
	loggerPath := helper.LogfilePath(fmt.Sprintf("%s-%d.log", prefix, e.ID))
	e.ObservedLogs, e.Log = helper.NewTestLoggerWithPath(loggerPath)
	crlog.SetLogger(zapr.NewLogger(e.Log.Desugar()))

	return ctxlog.NewParentContext(e.Log)
}

// FlushLog flushes the zap log
func (e *Environment) FlushLog() error {
	return e.Log.Sync()
}

// AllLogMessages returns only the message part of existing logs to aid in debugging
func (e *Environment) AllLogMessages() (msgs []string) {
	for _, m := range e.ObservedLogs.All() {
		msgs = append(msgs, m.Message)
	}

	return
}
