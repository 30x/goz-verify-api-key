package go_gateway_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoGateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoGateway Suite")
}
