package adapter_test

import (
	"github.com/df010/redis-service-adapter/adapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

var _ = Describe("generating manifests", func() {
	boolPointer := func(b bool) *bool {
		return &b
	}

	var (
		serviceRelease = serviceadapter.ServiceRelease{
			Name:    "wicked-release",
			Version: "4",
			Jobs:    []string{"redis"},
		}
		serviceDeployment = serviceadapter.ServiceDeployment{
			DeploymentName: "a-great-deployment",
			Releases:       serviceadapter.ServiceReleases{serviceRelease},
			Stemcell:       serviceadapter.Stemcell{OS: "TempleOS", Version: "4.05"},
		}
		generatedInstanceGroups = []bosh.InstanceGroup{{
			Name:     "redis",
			Networks: []bosh.Network{{Name: "an-etwork"}},
			Jobs: []bosh.Job{{
				Name: "redis",
			}},
		}}
		planInstanceGroups = []serviceadapter.InstanceGroup{
			{
				Name: "redis",
			},
		}
		previousPlanInstanceGroups = []serviceadapter.InstanceGroup{}

		manifest    bosh.BoshManifest
		generateErr error
		plan        serviceadapter.Plan

		actualInstanceGroups  []serviceadapter.InstanceGroup
		actualServiceReleases serviceadapter.ServiceReleases
		actualStemcell        string
	)

	BeforeEach(func() {
		plan = serviceadapter.Plan{InstanceGroups: planInstanceGroups}
	})
	BeforeEach(func() {
		adapter.InstanceGroupMapper = func(instanceGroups []serviceadapter.InstanceGroup,
			serviceReleases serviceadapter.ServiceReleases,
			stemcell string,
			deploymentInstanceGroupsToJobs map[string][]string) ([]bosh.InstanceGroup, error) {
			actualInstanceGroups = instanceGroups
			actualServiceReleases = serviceReleases
			actualStemcell = stemcell
			return generatedInstanceGroups, nil
		}
	})

	Context("when the instance group mapper maps instance groups successfully", func() {
		JustBeforeEach(func() {
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, map[string]interface{}{}, nil, nil)
		})

		It("returns no error", func() {
			Expect(generateErr).NotTo(HaveOccurred())
		})

		It("returns the basic deployment information", func() {
			Expect(manifest.Name).To(Equal(serviceDeployment.DeploymentName))
			Expect(manifest.Releases).To(ConsistOf(bosh.Release{Name: serviceRelease.Name, Version: serviceRelease.Version}))
			stemcellAlias := manifest.Stemcells[0].Alias
			Expect(manifest.Stemcells).To(ConsistOf(bosh.Stemcell{Alias: stemcellAlias, OS: serviceDeployment.Stemcell.OS, Version: serviceDeployment.Stemcell.Version}))
		})

		It("returns the instance groups produced by the mapper", func() {
			Expect(manifest.InstanceGroups).To(Equal(generatedInstanceGroups))
		})

		It("passes expected parameters to instance group mapper", func() {
			Expect(actualInstanceGroups).To(Equal(planInstanceGroups))
			Expect(actualServiceReleases).To(ConsistOf(serviceRelease))
			Expect(actualStemcell).To(Equal(manifest.Stemcells[0].Alias))
		})

		It("adds some reasonable defaults for update", func() {
			Expect(manifest.Update.Canaries).To(Equal(1))
			Expect(manifest.Update.MaxInFlight).To(Equal(10))
			Expect(manifest.Update.CanaryWatchTime).To(Equal("30000-240000"))
			Expect(manifest.Update.UpdateWatchTime).To(Equal("30000-240000"))
			Expect(manifest.Update.Serial).To(Equal(boolPointer(false)))
		})
	})

	Context("plan migrations", func() {
		var (
			previousPlan serviceadapter.Plan
		)
		BeforeEach(func() {
			planInstanceGroups = []serviceadapter.InstanceGroup{{Name: "redis", Instances: 2}}
			plan = serviceadapter.Plan{InstanceGroups: planInstanceGroups}
		})
		JustBeforeEach(func() {
			previousPlan = serviceadapter.Plan{InstanceGroups: previousPlanInstanceGroups}
			stderr.Reset()
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, map[string]interface{}{"parameters": map[string]interface{}{}}, nil, &previousPlan)
		})

		Context("when the previous plan had more redis", func() {
			BeforeEach(func() {
				previousPlanInstanceGroups = []serviceadapter.InstanceGroup{{Name: "redis", Instances: 3}}
			})
			It("fails", func() {
				Expect(generateErr).To(HaveOccurred())
			})
			It("logs details of the error for the operator", func() {
				Expect(stderr.String()).To(ContainSubstring("cannot migrate to a smaller plan"))
			})
			It("returns an empty error to the cli user", func() {
				Expect(generateErr.Error()).To(Equal(""))
			})
		})
		Context("when the previous plan had same number of redis", func() {
			BeforeEach(func() {
				previousPlanInstanceGroups = []serviceadapter.InstanceGroup{{Name: "redis", Instances: 2}}
			})
			It("succeeds", func() {
				Expect(generateErr).NotTo(HaveOccurred())
			})
			It("does not log an error for the operator", func() {
				Expect(stderr.String()).To(Equal(""))
			})
		})
		Context("when the previous plan had fewer redis", func() {
			BeforeEach(func() {
				previousPlanInstanceGroups = []serviceadapter.InstanceGroup{{Name: "redis", Instances: 1}}
			})
			It("succeeds", func() {
				Expect(generateErr).NotTo(HaveOccurred())
			})
			It("does not log an error for the operator", func() {
				Expect(stderr.String()).To(Equal(""))
			})
		})
	})
})
