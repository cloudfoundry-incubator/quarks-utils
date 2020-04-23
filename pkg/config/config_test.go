package config_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

var _ = Describe("Shared config", func() {
	Describe("constructors", func() {
		It("sets only meltdown defaults from constants, without manipulating arguments", func() {
			for _, fs := range []afero.Fs{nil, afero.NewOsFs()} {
				c := config.NewDefaultConfig(fs)
				Expect(c.OperatorNamespace).To(BeEmpty())
				Expect(c.WebhookServerHost).To(BeEmpty())
				Expect(c.CtxTimeOut).To(Equal(0 * time.Second))
				Expect(c.WebhookUseServiceRef).To(BeFalse())
				Expect(c.WebhookServerPort).To(Equal(int32(0)))
				// Check there is no manipulation with fs
				if fs == nil {
					Expect(c.Fs).To(BeNil())
				} else {
					Expect(c.Fs).To(Equal(fs))
				}
				Expect(c.MaxBoshDeploymentWorkers).To(Equal(0))
				Expect(c.MaxQuarksJobWorkers).To(Equal(0))
				Expect(c.MaxQuarksSecretWorkers).To(Equal(0))
				Expect(c.MaxQuarksStatefulSetWorkers).To(Equal(0))
				Expect(c.MeltdownDuration).To(Equal(config.MeltdownDuration))
				Expect(c.MeltdownRequeueAfter).To(Equal(config.MeltdownRequeueAfter))
			}
		})
	})
})
