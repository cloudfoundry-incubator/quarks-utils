package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

var _ = Describe("Docker image", func() {
	Describe("reference", func() {
		It("returns the location of the docker image", func() {
			err := config.SetupOperatorDockerImage("foo", "bar", "1.2.3")
			Expect(err).ToNot(HaveOccurred())
			Expect(config.GetOperatorDockerImage()).To(Equal("foo/bar:1.2.3"))
		})
	})

	Describe("pull policy", func() {
		Context("when policy is invalid", func() {
			It("returns an error", func() {
				err := config.SetupOperatorImagePullPolicy("fake-policy")
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when policy is empty", func() {
			It("returns an error", func() {
				err := config.SetupOperatorImagePullPolicy("")
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when policy is valid", func() {
			It("returns an error", func() {
				for _, policy := range []corev1.PullPolicy{corev1.PullAlways, corev1.PullIfNotPresent, corev1.PullNever} {
					err := config.SetupOperatorImagePullPolicy(string(policy))
					Expect(err).ToNot(HaveOccurred())
					Expect(config.GetOperatorImagePullPolicy()).To(Equal(policy))
				}
			})
		})
	})
})
