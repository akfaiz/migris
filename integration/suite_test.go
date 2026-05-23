package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo specs use dot imports by convention.
	. "github.com/onsi/gomega"    //nolint:revive // Ginkgo specs use dot imports by convention.
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
