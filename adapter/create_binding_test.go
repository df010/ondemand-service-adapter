package adapter_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

var _ = Describe("Create", func() {
	os.Remove("/var/vcap/jobs/ondemand/config/config.yml")
	ioutil.WriteFile("/var/vcap/jobs/ondemand/config/config.yml", []byte(configString), 0644)

	var (
		boshVMs           map[string][]string
		binding           serviceadapter.Binding
		requestParameters map[string]interface{}
		bindErr           error
	)

	JustBeforeEach(func() {
		binding, bindErr = binder.CreateBinding("binding-id", boshVMs, bosh.BoshManifest{
			Properties: map[string]interface{}{"user": map[string]interface{}{"password": "pass"},
				"keepalived": map[string]interface{}{"vip": "192.168.1.1"},
			},
		}, requestParameters)
	})

	Context("when there are vms", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{"haproxy": {"192.168.0.1", "192.168.0.2"}}
		})

		It("returns no error", func() {
			Expect(bindErr).NotTo(HaveOccurred())
		})

		It("returns prop credentials", func() {
			Expect(binding).To(Equal(
				serviceadapter.Binding{
					Credentials: map[string]interface{}{"ip": []string{"192.168.0.1", "192.168.0.2"},
						"vip": "192.168.1.1", "user": "cloud", "password": "pass"},
				},
			))
		})
	})

	Context("when there are no vms for redis", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{}
		})

		It("returns an error ", func() {
			Expect(bindErr).To(HaveOccurred())
		})

		// It("logs an error for the operator", func() {
		// 	Expect(stderr.String()).To(ContainSubstring("no VMs for instance group redis"))
		// })
	})
})
