package c_gateway_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCGateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CGateway Suite")
}
