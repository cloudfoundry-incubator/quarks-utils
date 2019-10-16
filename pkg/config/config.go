package config

import (
	"time"

	"github.com/spf13/afero"
)

const (
	// MeltdownDuration is the duration of the meltdown period, in which we
	// postpone further reconciles for the same resource
	MeltdownDuration = 10 * time.Second
	// MeltdownRequeueAfter is the duration for which we delay the requeuing of the reconcile
	MeltdownRequeueAfter = 5 * time.Second
)

// Config controls the behaviour of different controllers
type Config struct {
	CtxTimeOut                    time.Duration
	MeltdownDuration              time.Duration
	MeltdownRequeueAfter          time.Duration
	Namespace                     string
	OperatorNamespace             string
	WebhookUseServiceRef          bool
	WebhookServerHost             string
	WebhookServerPort             int32
	Fs                            afero.Fs
	MaxBoshDeploymentWorkers      int
	MaxExtendedJobWorkers         int
	MaxExtendedSecretWorkers      int
	MaxExtendedStatefulSetWorkers int
}

// NewDefaultConfig returns a new Config for a manager of controllers
func NewDefaultConfig(fs afero.Fs) *Config {
	return &Config{
		MeltdownDuration:     MeltdownDuration,
		MeltdownRequeueAfter: MeltdownRequeueAfter,
		Fs:                   fs,
	}
}

// NewConfig returns a new Config for a manager of controllers
func NewConfig(namespace string, operatorNamespace string, ctxTimeOut int, useServiceRef bool, host string, port int32, fs afero.Fs, maxBoshDeploymentWorkers, maxExtendedJobWorkers, maxExtendedSecretWorkers, maxExtendedStatefulSetWorkers int) *Config {
	return &Config{
		CtxTimeOut:                    time.Duration(ctxTimeOut) * time.Second,
		MeltdownDuration:              MeltdownDuration,
		MeltdownRequeueAfter:          MeltdownRequeueAfter,
		Namespace:                     namespace,
		OperatorNamespace:             operatorNamespace,
		WebhookUseServiceRef:          useServiceRef,
		WebhookServerHost:             host,
		WebhookServerPort:             port,
		Fs:                            fs,
		MaxBoshDeploymentWorkers:      maxBoshDeploymentWorkers,
		MaxExtendedJobWorkers:         maxExtendedJobWorkers,
		MaxExtendedSecretWorkers:      maxExtendedSecretWorkers,
		MaxExtendedStatefulSetWorkers: maxExtendedStatefulSetWorkers,
	}
}
