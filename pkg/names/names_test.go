package names_test

import (
	"fmt"

	"code.cloudfoundry.org/quarks-utils/pkg/names"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Names", func() {
	type test struct {
		arg    string
		result string
		n      int
	}
	str31 := "a123456789012345678901234567890"
	str32 := `123456789012345678901234567890AA`
	str63 := str31 + "b123456789012345678901234567890c"
	str253 := str63 + str63 + str63 + str63 + "d"

	Context("JobName", func() {
		tests := []test{
			{arg: "ab1", result: "ab1", n: 20},
			{arg: "a-b1", result: "a-b1-", n: 21},
			{arg: str31, result: "a123456789012345678901234567890-", n: 48},
			{arg: str63, result: str63[:39] + "-", n: 56},
		}

		It("produces valid k8s job names", func() {
			for _, t := range tests {
				r, err := names.JobName(t.arg)
				Expect(err).ToNot(HaveOccurred())
				Expect(r).To(ContainSubstring(t.result), fmt.Sprintf("%#v", t))
				Expect(r).To(HaveLen(t.n), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("Sanitize", func() {
		// according to docs/naming.md
		tests := []test{
			{arg: "AB1", result: "ab1"},
			{arg: "ab1", result: "ab1"},
			{arg: "1bc", result: "1bc"},
			{arg: "a-b", result: "a-b"},
			{arg: "a_b", result: "a-b"},
			{arg: "a_b_123", result: "a-b-123"},
			{arg: "-abc", result: "abc"},
			{arg: "abc-", result: "abc"},
			{arg: "_abc_", result: "abc"},
			{arg: "-abc-", result: "abc"},
			{arg: "abcü.123:4", result: "abc1234"},
			{arg: str63, result: str63},
			{arg: str63 + "0", result: str31[:30] + "-f61acdbce0e8ea6e4912f53bde4de866"},
		}

		It("produces valid k8s names", func() {
			for _, t := range tests {
				Expect(names.Sanitize(t.arg)).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("SanitizeSubdomain", func() {
		tests := []test{
			{arg: "AB1", result: "ab1"},
			{arg: "ab1", result: "ab1"},
			{arg: "1bc", result: "1bc"},
			{arg: "a-b", result: "a-b"},
			{arg: "a_b", result: "a-b"},
			{arg: "a_b_123", result: "a-b-123"},
			{arg: "-abc", result: "abc"},
			{arg: "abc-", result: "abc"},
			{arg: "_abc_", result: "abc"},
			{arg: "-abc-", result: "abc"},
			{arg: ".a.b.c.", result: "a.b.c"},
			{arg: "abcü.123:4", result: "abc.1234"},
			{arg: str253, result: str253},
		}

		It("produces valid k8s names", func() {
			for _, t := range tests {
				Expect(names.SanitizeSubdomain(t.arg)).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})

		It("shortens long names", func() {
			Expect(names.SanitizeSubdomain(str253 + "1")).To(HaveLen(253))
		})
	})

	Context("VolumeName", func() {
		tests := []test{
			{arg: "secret", result: "secret"},
			{arg: "secret.name", result: "name"},
			{arg: str63, result: str63},
			{arg: str63 + ".foo", result: "foo"},
			{arg: "foo." + str63, result: "a123456789012345678901234567890b123456789012345678901234567890c"},
		}

		It("produces valid k8s names", func() {
			for _, t := range tests {
				Expect(names.VolumeName(t.arg)).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("DNSLabelSafe", func() {
		tests := []test{
			{arg: "-Name", result: "name"},
			{arg: "--_Name--", result: "name"},
			{arg: "nA!müe123", result: "name123"},
		}

		It("produces valid k8s DNS labels chars", func() {
			for _, t := range tests {
				r := names.DNSLabelSafe(t.arg)
				Expect(r).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})
	})

	Context("TruncateMD5", func() {
		const (
			longstringMD5 = `d3b442d33a939d736f8d44a1fbefa679`
		)

		tests := []test{
			{arg: "longstring", n: 3, result: longstringMD5[:3]},
			{arg: "longstring", n: 10, result: "longstring"},
			{arg: "longstring", n: 11, result: "longstring"},
			{
				arg: "longstring" + str32, n: 32,
				result: "4fa56ac8cd8f1badfb3226365b0771e8",
			},
			{
				arg: "longstring" + str32, n: 33,
				result: "4fa56ac8cd8f1badfb3226365b0771e8",
			},
			{
				arg: "longstringAA" + str32, n: 43,
				result: "longstring-07b4553eabfa74c177f570f0b87ce33e",
			},
			{
				arg: "longstring" + str32 + str32, n: 63,
				result: "longstring12345678901234567890-e42ccf2f5a0b4aca30676ffa42b15d16",
			},
		}

		It("produces valid k8s DNS labels chars", func() {
			for _, t := range tests {
				r := names.TruncateMD5(t.arg, t.n)
				Expect(r).To(Equal(t.result), fmt.Sprintf("%#v", t))
			}
		})
	})
})
