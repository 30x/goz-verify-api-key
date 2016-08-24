package verifyAPIKey_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVerifyAPIKey(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VerifyAPIKey Suite")
}
