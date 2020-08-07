package inmemorygenerator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/quarks-utils/pkg/credsgen"
	inmemorygenerator "code.cloudfoundry.org/quarks-utils/pkg/credsgen/in_memory_generator"
	helper "code.cloudfoundry.org/quarks-utils/testing/testhelper"
)

var _ = Describe("InMemoryGenerator", func() {
	var (
		generator credsgen.Generator
	)

	BeforeEach(func() {
		_, log := helper.NewTestLogger()
		generator = inmemorygenerator.NewInMemoryGenerator(log)
	})

	Describe("GenerateSSHKey", func() {
		It("generates an SSH key", func() {
			key, err := generator.GenerateSSHKey("foo")

			Expect(err).ToNot(HaveOccurred())
			Expect(key.PrivateKey).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
			Expect(key.PublicKey).To(MatchRegexp("ssh-rsa\\s.+"))
			Expect(key.Fingerprint).To(MatchRegexp("([0-9a-f]{2}:){15}[0-9a-f]{2}"))
		})
	})
})
