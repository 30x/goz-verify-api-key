package mockhttpserver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMockhttpserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mockhttpserver Suite")
}
