package meltdown_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeltdown(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meltdown Suite")
}
