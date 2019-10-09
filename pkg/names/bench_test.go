// go test -bench=. -benchmem ./pkg/kube/util/names
package names_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

var (
	// max = 100
	// all = 201
	max = 5
	all = 11
)

func BenchmarkFlatten(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := make([]corev1.Container, 0)
		t := make([]corev1.Container, 0)

		for i := 0; i < max; i++ {
			s = append(s, container())
		}
		for i := 0; i < max; i++ {
			t = append(t, container())
		}

		r := flattenContainers(s, container(), t)
		if len(r) != all {
			b.FailNow()
		}

	}
}

func BenchmarkFlattenNoMake(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := []corev1.Container{}
		t := []corev1.Container{}

		for i := 0; i < max; i++ {
			s = append(s, container())
		}
		for i := 0; i < max; i++ {
			t = append(t, container())
		}

		r := flattenContainers(s, container(), t)
		if len(r) != all {
			b.FailNow()
		}

	}
}

func BenchmarkSimpleAppend(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := []corev1.Container{}
		t := []corev1.Container{}

		for i := 0; i < max; i++ {
			s = append(s, container())
		}
		for i := 0; i < max; i++ {
			t = append(t, container())
		}

		r := []corev1.Container{}
		r = append(r, s...)
		r = append(r, container())
		r = append(r, t...)
		if len(r) != all {
			b.FailNow()
		}
	}
}

func BenchmarkProvideCap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := make([]corev1.Container, 0, 10)
		t := make([]corev1.Container, 0, 10)

		for i := 0; i < max; i++ {
			s = append(s, container())
		}
		for i := 0; i < max; i++ {
			t = append(t, container())
		}

		r := make([]corev1.Container, 0, 10)
		r = append(r, s...)
		r = append(r, container())
		r = append(r, t...)
		if len(r) != all {
			b.FailNow()
		}
	}
}

// flattenContainers will flatten the containers parameter. Each argument passed to
// flattenContainers should be a corev1.Container or []corev1.Container. The final
// []corev1.Container creation is optimized to prevent slice re-allocation.
func flattenContainers(containers ...interface{}) []corev1.Container {
	var totalLen int
	for _, instance := range containers {
		switch v := instance.(type) {
		case []corev1.Container:
			totalLen += len(v)
		case corev1.Container:
			totalLen++
		}
	}
	result := make([]corev1.Container, 0, totalLen)
	for _, instance := range containers {
		switch v := instance.(type) {
		case []corev1.Container:
			result = append(result, v...)
		case corev1.Container:
			result = append(result, v)
		}
	}
	return result
}

func container() corev1.Container {
	return corev1.Container{
		Name:    "test",
		Image:   "test",
		Command: []string{"/bin/sh", "-c"},
		Args:    []string{"test"},
	}
}
