package adapter_test

import (
	"bytes"
	"io"
	"log"

	"github.com/df010/ondemand-service-adapter/adapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	manifestGenerator *adapter.ManifestGenerator
	binder            *adapter.Binder
	stderr            bytes.Buffer
	configString      string = "" +
		`---
binding_credentials:
- name: haproxy
  plan: haproxy
  datatype: array
  value: "[JOB.haproxy.ip]"
input_mappings:
- key: keepalived.vip1
  valueformat: ip_range
  valuemap: 1:all
- key: keepalived.vip2
  valueformat: ip_range
  valuemap: 1:all
- key: keepalived.vip3
  valueformat: ip_range
  valuemap: 1:all
- key: keepalived.virtual_router_id
  valueformat: number_range
  valuemap: 2:all
- key: keepalived.priority
  valueformat: number
  valuemap: "1"
- key: keepalived.test
  valueformat: number
  valuemap: "any:all"
- key: keepalived.test
  valueformat: number
  valuemap: "any:all"
instance_groups:
- name: haproxy
  templates:
  - haproxy
    keepalived`
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
