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
				Expect(c.Namespace).To(BeEmpty())
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

		It("properly sets a shared config", func() {
			fs := afero.NewOsFs()

			c := config.NewConfig(
				"test",
				"operator-namespace",
				10,
				true,
				"0.0.0.0",
				8080,
				fs,
				1, 2, 3, 4, 5, 6)

			Expect(c.Namespace).To(Equal("test"))
			Expect(c.OperatorNamespace).To(Equal("operator-namespace"))
			Expect(c.WebhookServerHost).To(Equal("0.0.0.0"))
			Expect(c.CtxTimeOut).To(Equal(10 * time.Second))
			Expect(c.WebhookUseServiceRef).To(BeTrue())
			Expect(c.WebhookServerPort).To(Equal(int32(8080)))
			Expect(c.Fs).To(Equal(fs))
			Expect(c.MaxBoshDeploymentWorkers).To(Equal(1))
			Expect(c.MaxQuarksJobWorkers).To(Equal(2))
			Expect(c.MaxQuarksSecretWorkers).To(Equal(3))
			Expect(c.MaxQuarksStatefulSetWorkers).To(Equal(4))
			Expect(c.MeltdownDuration).To(Equal(5 * time.Second))
			Expect(c.MeltdownRequeueAfter).To(Equal(6 * time.Second))
		})

	})
})
