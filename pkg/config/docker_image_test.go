package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

var _ = Describe("config", func() {
	Describe("GetOperatorDockerImage", func() {
		It("returns the location of the docker image", func() {
			err := config.SetupOperatorDockerImage("foo", "bar", "1.2.3")
			Expect(err).ToNot(HaveOccurred())
			Expect(config.GetOperatorDockerImage()).To(Equal("foo/bar:1.2.3"))
		})
	})
})
