package names_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

var _ = Describe("SpecIndex", func() {
	tests := []struct {
		az  int
		ord int
		r   int
	}{
		{az: 0, ord: 0, r: 0},
		{az: 1, ord: 5, r: 5},
		{az: 2, ord: 5, r: 10005},
		{az: 3, ord: 5, r: 20005},
		{az: -1, ord: 0, r: 0},
	}

	It("produces valid spec indexes", func() {
		for _, t := range tests {
			r := names.SpecIndex(t.az, t.ord)
			Expect(r).To(Equal(t.r), fmt.Sprintf("%#v", t))
		}
	})

})
