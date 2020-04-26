package config

import (
	"time"

	"github.com/spf13/afero"
)

const (
	// MeltdownDuration is the duration of the meltdown period, in which we
	// postpone further reconciles for the same resource
	MeltdownDuration = 1 * time.Minute
	// MeltdownRequeueAfter is the duration for which we delay the requeuing of the reconcile
	MeltdownRequeueAfter = 30 * time.Second
)

// Config controls the behaviour of different controllers
type Config struct {
	CtxTimeOut           time.Duration
	MeltdownDuration     time.Duration
	MeltdownRequeueAfter time.Duration
	// Namespace that is being watched by controllers
	Namespace string
	// OperatorNamespace is where the webhook services of the operator are placed
	OperatorNamespace           string
	WebhookUseServiceRef        bool
	WebhookServerHost           string
	WebhookServerPort           int32
	Fs                          afero.Fs
	MaxBoshDeploymentWorkers    int
	MaxQuarksJobWorkers         int
	MaxQuarksSecretWorkers      int
	MaxQuarksStatefulSetWorkers int
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
func NewConfig(namespace string, operatorNamespace string, ctxTimeOut int, useServiceRef bool, host string, port int32, fs afero.Fs, maxBoshDeploymentWorkers, maxQuarksJobWorkers, maxQuarksSecretWorkers, maxQuarksStatefulSetWorkers int) *Config {
	return &Config{
		CtxTimeOut:                  time.Duration(ctxTimeOut) * time.Second,
		MeltdownDuration:            MeltdownDuration,
		MeltdownRequeueAfter:        MeltdownRequeueAfter,
		Namespace:                   namespace,
		OperatorNamespace:           operatorNamespace,
		WebhookUseServiceRef:        useServiceRef,
		WebhookServerHost:           host,
		WebhookServerPort:           port,
		Fs:                          fs,
		MaxBoshDeploymentWorkers:    maxBoshDeploymentWorkers,
		MaxQuarksJobWorkers:         maxQuarksJobWorkers,
		MaxQuarksSecretWorkers:      maxQuarksSecretWorkers,
		MaxQuarksStatefulSetWorkers: maxQuarksStatefulSetWorkers,
	}
}
