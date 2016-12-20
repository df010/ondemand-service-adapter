package adapter_test

import (
	"bytes"
	"io"
	"log"

	"github.com/df010/redis-service-adapter/adapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	manifestGenerator *adapter.ManifestGenerator
	binder            *adapter.Binder
	stderr            bytes.Buffer
)

var _ = BeforeEach(func() {
	manifestGenerator = &adapter.ManifestGenerator{
		StderrLogger: log.New(io.MultiWriter(GinkgoWriter, &stderr), "", log.LstdFlags),
	}
	binder = &adapter.Binder{
		StderrLogger: log.New(io.MultiWriter(GinkgoWriter, &stderr), "", log.LstdFlags),
	}
})

func TestAdapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Adapter Suite")
}
