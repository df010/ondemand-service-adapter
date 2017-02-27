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
		boshVMs map[string][]string

		binding           serviceadapter.Binding
		requestParameters map[string]interface{}
		bindErr           error
	)

	JustBeforeEach(func() {
		binding, bindErr = binder.CreateBinding("binding-id", boshVMs, bosh.BoshManifest{}, requestParameters)
	})

	Context("when there are vms", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{"redis": {"foo", "bar"}}
		})

		It("returns no error", func() {
			Expect(bindErr).NotTo(HaveOccurred())
		})

		It("returns redis in credentials", func() {
			Expect(binding).To(Equal(
				serviceadapter.Binding{
					Credentials: map[string]interface{}{"redis": []interface{}{"foo:6379", "bar:6379"}},
				},
			))
		})
	})

	Context("when there are no vms for redis", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{"baz": {"foo", "bar"}}
		})

		It("returns an error to the cli user", func() {
			Expect(bindErr).To(MatchError(""))
		})

		It("logs an error for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("no VMs for instance group redis"))
		})
	})
})
