package adapter_test

import (
	"github.com/df010/ondemand-service-adapter/adapter"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Adapter/GenerateDashboardUrl", func() {
	It("generates a arbitrary dashboard url", func() {
		generator := &adapter.DashboardUrlGenerator{}
		//Expect(generator.DashboardUrl("instanceID", serviceadapter.Plan{}, bosh.BoshManifest{})).To(Equal(serviceadapter.DashboardUrl{DashboardUrl: "http://example_dashboard.com/instanceID"}))
		Expect(generator.DashboardUrl("instanceID", serviceadapter.Plan{}, bosh.BoshManifest{})).To(Equal(serviceadapter.DashboardUrl{DashboardUrl: ""}))
	})
})
