package names_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

var _ = Describe("Names", func() {
	type test struct {
		arg1   string
		result string
		n      int
	}
	long31 := "a123456789012345678901234567890"
	long63 := long31 + "b123456789012345678901234567890c"
	long253 := long63 + long63 + long63 + long63 + "d"

	Context("JobName", func() {
		tests := []test{
			{arg1: "ab1", result: "ab1", n: 20},
			{arg1: "a-b1", result: "a-b1-", n: 21},
			{arg1: long31, result: "a123456789012345678901234567890-", n: 48},
			{arg1: long63, result: long63[:39] + "-", n: 56},
		}

		It("produces valid k8s job names", func() {
			for _, t := range tests {
				r, err := names.JobName(t.arg1)
				Expect(err).ToNot(HaveOccurred())
				Expect(r).To(ContainSubstring(t.result), fmt.Sprintf("%#v", t))
				Expect(r).To(HaveLen(t.n), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("Sanitize", func() {
		// according to docs/naming.md
		tests := []test{
			{arg1: "AB1", result: "ab1"},
			{arg1: "ab1", result: "ab1"},
			{arg1: "1bc", result: "1bc"},
			{arg1: "a-b", result: "a-b"},
			{arg1: "a_b", result: "a-b"},
			{arg1: "a_b_123", result: "a-b-123"},
			{arg1: "-abc", result: "abc"},
			{arg1: "abc-", result: "abc"},
			{arg1: "_abc_", result: "abc"},
			{arg1: "-abc-", result: "abc"},
			{arg1: "abcü.123:4", result: "abc1234"},
			{arg1: long63, result: long63},
			{arg1: long63 + "0", result: long31 + "f61acdbce0e8ea6e4912f53bde4de866"},
		}

		It("produces valid k8s names", func() {
			for _, t := range tests {
				Expect(names.Sanitize(t.arg1)).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("SanitizeSubdomain", func() {
		tests := []test{
			{arg1: "AB1", result: "ab1"},
			{arg1: "ab1", result: "ab1"},
			{arg1: "1bc", result: "1bc"},
			{arg1: "a-b", result: "a-b"},
			{arg1: "a_b", result: "a-b"},
			{arg1: "a_b_123", result: "a-b-123"},
			{arg1: "-abc", result: "abc"},
			{arg1: "abc-", result: "abc"},
			{arg1: "_abc_", result: "abc"},
			{arg1: "-abc-", result: "abc"},
			{arg1: ".a.b.c.", result: "a.b.c"},
			{arg1: "abcü.123:4", result: "abc.1234"},
			{arg1: long253, result: long253},
		}

		It("produces valid k8s names", func() {
			for _, t := range tests {
				Expect(names.SanitizeSubdomain(t.arg1)).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})

		It("shortens long names", func() {
			Expect(names.SanitizeSubdomain(long253 + "1")).To(HaveLen(253))
		})
	})

	Context("VolumeName", func() {
		tests := []test{
			{arg1: "secret", result: "secret"},
			{arg1: "secret.name", result: "name"},
			{arg1: long63, result: long63},
			{arg1: long63 + ".foo", result: "foo"},
			{arg1: "foo." + long63, result: "a123456789012345678901234567890b123456789012345678901234567890c"},
		}

		It("produces valid k8s names", func() {
			for _, t := range tests {
				Expect(names.VolumeName(t.arg1)).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("SecretName", func() {
		type test struct {
			arg1   names.DeploymentSecretType
			arg2   string
			arg3   string
			arg4   string
			name   string
			prefix string
		}
		tests := []test{
			{
				arg1:   names.DeploymentSecretTypeInstanceGroupResolvedProperties,
				arg2:   "deploymentName",
				arg3:   "ig-Name",
				arg4:   "", // "0.1",
				name:   "deployment.ig-resolved.ig-name",
				prefix: "deployment.ig-resolved.ig-name",
			},
			{
				arg1:   names.DeploymentSecretTypeInstanceGroupResolvedProperties,
				arg2:   "deployment-Name",
				arg3:   "ig_Name",
				arg4:   "",
				name:   "deployment-.ig-resolved.ig-name",
				prefix: "deployment-.ig-resolved.ig-name",
			},
			{
				arg1:   names.DeploymentSecretTypeInstanceGroupResolvedProperties,
				arg2:   "deploymentname123456789012345678901234567890123456789012345678901234567890",
				arg3:   "igname1234567890123456789012345678901234567890123456789012345678901234567890",
				arg4:   "",
				name:   "deploymentname123456789012345679b45c361de1db4171f6a0ea0bbe035ed.igname1234567890123456789012345ac27e305a5b88c7c20380c2737917bbc",
				prefix: "deploymentname123456789012345679b45c361de1db4171f6a0ea0bbe035ed.igname1234567890123456789012345ac27e305a5b88c7c20380c2737917bbc",
			},
		}

		It("produces valid k8s secret names", func() {
			for _, t := range tests {
				r := names.InstanceGroupSecretName(t.arg1, t.arg2, t.arg3, t.arg4)
				Expect(r).To(Equal(t.name), fmt.Sprintf("%#v", t))
			}
		})

		It("produces valid prefixes", func() {
			for _, t := range tests {
				r := names.DeploymentSecretPrefix(t.arg1, t.arg2) + names.Sanitize(t.arg3)
				Expect(r).To(Equal(t.prefix), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("QuarksLink related names", func() {
		Context("link type and name strings", func() {
			It("should return a simple link type and link name string", func() {
				Expect(names.QuarksLinkSecretKey("type", "name")).To(Equal("type-name"))
			})
		})

		Context("link secret name strings", func() {
			It("should return a string to be used as a secret name for links without a suffix", func() {
				Expect(names.QuarksLinkSecretName("deploymentname")).To(Equal("link-deploymentname"))
			})

			It("should return a string to be used as a secret name for links with one suffix", func() {
				Expect(names.QuarksLinkSecretName("deploymentname", "suffix")).To(Equal("link-deploymentname-suffix"))
			})

			It("should return a string to be used as a secret name for links with multiple suffixes", func() {
				Expect(names.QuarksLinkSecretName("deploymentname", "one", "two")).To(Equal("link-deploymentname-one-two"))
			})
		})
	})
})
