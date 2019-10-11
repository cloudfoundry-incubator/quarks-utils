package meltdown_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"code.cloudfoundry.org/quarks-utils/pkg/meltdown"
)

var _ = Describe("Meltdown", func() {

	const MeltdownDuration time.Duration = 10 * time.Second

	Describe("NewAnnotationWindow", func() {
		annotation := func(t time.Time) map[string]string {
			return map[string]string{
				meltdown.AnnotationLastReconcile: t.Format(time.RFC3339),
			}
		}

		It("returns true if we're in active meltdown", func() {
			start := time.Now()
			Expect(meltdown.NewAnnotationWindow(MeltdownDuration, annotation(start)).Contains(start)).To(BeTrue())
			now := start.Add(MeltdownDuration - 1*time.Second)
			Expect(meltdown.NewAnnotationWindow(MeltdownDuration, annotation(start)).Contains(now)).To(BeTrue())
		})

		It("returns false if we're outside the active meltdown", func() {
			start := time.Now()
			end := start.Add(MeltdownDuration)
			Expect(meltdown.NewAnnotationWindow(MeltdownDuration, annotation(start)).Contains(end)).To(BeFalse())

			before := start.Add(-1 * time.Second)
			Expect(meltdown.NewAnnotationWindow(MeltdownDuration, annotation(start)).Contains(before)).To(BeFalse())
		})

		It("returns false if we have no last reconcile timestamp", func() {
			end := time.Now()
			Expect(meltdown.NewAnnotationWindow(MeltdownDuration, map[string]string{}).Contains(end)).To(BeFalse())
		})
	})

	Describe("NewWindow", func() {
		It("returns true if we're in active meltdown", func() {
			start := time.Now()
			lastReconcile := metav1.NewTime(start)
			Expect(meltdown.NewWindow(MeltdownDuration, &lastReconcile).Contains(start)).To(BeTrue())

			now := start.Add(MeltdownDuration - 1*time.Second)
			Expect(meltdown.NewWindow(MeltdownDuration, &lastReconcile).Contains(now)).To(BeTrue())
		})

		It("returns false if we're outside the active meltdown", func() {
			start := time.Now()
			end := start.Add(MeltdownDuration)
			lastReconcile := metav1.NewTime(start)
			Expect(meltdown.NewWindow(MeltdownDuration, &lastReconcile).Contains(end)).To(BeFalse())

			before := start.Add(-1 * time.Second)
			Expect(meltdown.NewWindow(MeltdownDuration, &lastReconcile).Contains(before)).To(BeFalse())
		})

		It("returns false if we have no last reconcile timestamp", func() {
			end := time.Now()
			Expect(meltdown.NewWindow(MeltdownDuration, nil).Contains(end)).To(BeFalse())
		})
	})
})
