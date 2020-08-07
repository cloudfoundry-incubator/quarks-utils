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
		defaultGenerator credsgen.Generator
	)

	BeforeEach(func() {
		_, log := helper.NewTestLogger()
		defaultGenerator = inmemorygenerator.NewInMemoryGenerator(log)
	})

	Describe("NewInMemoryGenerator", func() {
		Context("object defaults", func() {
			It("succeeds if the default type is inmemorygenerator.InMemoryGenerator", func() {
				t, ok := defaultGenerator.(*inmemorygenerator.InMemoryGenerator)
				Expect(ok).To(BeTrue())
				Expect(t).To(Equal(defaultGenerator))
			})

			It("succeeds if the default generator is 2048 bits", func() {
				Expect(defaultGenerator.(*inmemorygenerator.InMemoryGenerator).Bits).To(Equal(2048))
			})

			It("succeeds if the default generator is rsa", func() {
				Expect(defaultGenerator.(*inmemorygenerator.InMemoryGenerator).Algorithm).To(Equal("rsa"))
			})

			It("succeeds if the default generator certs expires in 365 days", func() {
				Expect(defaultGenerator.(*inmemorygenerator.InMemoryGenerator).Expiry).To(Equal(365))
			})
		})
	})
})
