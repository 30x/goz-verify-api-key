package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGozVerifyApiKey(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GozVerifyApiKey Suite")
}
