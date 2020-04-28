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
				sharedConfig := config.NewDefaultConfig(fs)
				Expect(sharedConfig.Namespace).To(BeEmpty())
				Expect(sharedConfig.OperatorNamespace).To(BeEmpty())
				Expect(sharedConfig.WebhookServerHost).To(BeEmpty())
				Expect(sharedConfig.CtxTimeOut).To(Equal(0 * time.Second))
				Expect(sharedConfig.WebhookUseServiceRef).To(BeFalse())
				Expect(sharedConfig.WebhookServerPort).To(Equal(int32(0)))
				// Check there is no manipulation with fs
				if fs == nil {
					Expect(sharedConfig.Fs).To(BeNil())
				} else {
					Expect(sharedConfig.Fs).To(Equal(fs))
				}
				Expect(sharedConfig.MaxBoshDeploymentWorkers).To(Equal(0))
				Expect(sharedConfig.MaxQuarksJobWorkers).To(Equal(0))
				Expect(sharedConfig.MaxQuarksSecretWorkers).To(Equal(0))
				Expect(sharedConfig.MaxQuarksStatefulSetWorkers).To(Equal(0))
				Expect(sharedConfig.MeltdownDuration).To(Equal(config.MeltdownDuration))
				Expect(sharedConfig.MeltdownRequeueAfter).To(Equal(config.MeltdownRequeueAfter))
			}
		})
		It("properly sets a shared config", func() {
			fs := afero.NewOsFs()

			sharedConfig := config.NewConfig(
				"test",
				"operator-namespace",
				10,
				true,
				"0.0.0.0",
				8080,
				fs,
				1, 2, 3, 4, 5, 6)

			Expect(sharedConfig.Namespace).To(Equal("test"))
			Expect(sharedConfig.OperatorNamespace).To(Equal("operator-namespace"))
			Expect(sharedConfig.WebhookServerHost).To(Equal("0.0.0.0"))
			Expect(sharedConfig.CtxTimeOut).To(Equal(10 * time.Second))
			Expect(sharedConfig.WebhookUseServiceRef).To(BeTrue())
			Expect(sharedConfig.WebhookServerPort).To(Equal(int32(8080)))
			Expect(sharedConfig.Fs).To(Equal(fs))
			Expect(sharedConfig.MaxBoshDeploymentWorkers).To(Equal(1))
			Expect(sharedConfig.MaxQuarksJobWorkers).To(Equal(2))
			Expect(sharedConfig.MaxQuarksSecretWorkers).To(Equal(3))
			Expect(sharedConfig.MaxQuarksStatefulSetWorkers).To(Equal(4))
			Expect(sharedConfig.MeltdownDuration).To(Equal(5 * time.Second))
			Expect(sharedConfig.MeltdownRequeueAfter).To(Equal(6 * time.Second))
		})
	})
})
