package testhelper

import (
	"time"

	"github.com/spf13/afero"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// NewConfigWithTimeout returns a default config, with a context timeout
func NewConfigWithTimeout(timeout time.Duration) *config.Config {
	return &config.Config{
		Fs:                   afero.NewMemMapFs(),
		CtxTimeOut:           timeout,
		MeltdownDuration:     config.MeltdownDuration,
		MeltdownRequeueAfter: config.MeltdownRequeueAfter,
	}
}
