package adapter_test

import (
	"io/ioutil"
	"os"

	"github.com/df010/ondemand-service-adapter/adapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

var _ = Describe("generating manifests", func() {

	os.Remove("/var/vcap/jobs/ondemand/config/config.yml")
	ioutil.WriteFile("/var/vcap/jobs/ondemand/config/config.yml", []byte(configString), 0644)

	boolPointer := func(b bool) *bool {
		return &b
	}

	var (
		serviceRelease = serviceadapter.ServiceRelease{
			Name:    "wicked-release",
			Version: "4",
			Jobs:    []string{"haproxy", "keepalived"},
		}
		serviceDeployment = serviceadapter.ServiceDeployment{
			DeploymentName: "a-great-deployment",
			Releases:       serviceadapter.ServiceReleases{serviceRelease},
			Stemcell:       serviceadapter.Stemcell{OS: "TempleOS", Version: "4.05"},
		}
		generatedInstanceGroups = []bosh.InstanceGroup{{
			Name:     "haproxy",
			Networks: []bosh.Network{{Name: "an-etwork"}},
			Jobs: []bosh.Job{{
				Name: "haproxy",
			}, {
				Name: "keepalived",
			}},
		}}
		planInstanceGroups = []serviceadapter.InstanceGroup{
			{
				Name: "haproxy",
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

		// os.Remove("/var/vcap/jobs/ondemand/config/config.yml")
		// os.Link("/home/df/gocode/src/github.com/df010/ondemand-service-adapter/config.yml", "/var/vcap/jobs/ondemand/config/config.yml")
		plan = serviceadapter.Plan{InstanceGroups: planInstanceGroups, Properties: map[string]interface{}{"metadata_config": map[string]interface{}{"network": "a|b"}}}
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
			props := map[string]interface{}{"parameters": map[string]interface{}{}}
			props = props["parameters"].(map[string]interface{})
			props["metadata"] = map[string]interface{}{}
			props["metadata"].(map[string]interface{})["network"] = "an-etwork"

			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, props, nil, nil)
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

	Context("props inputs", func() {
		JustBeforeEach(func() {
			props := map[string]interface{}{"parameters": map[string]interface{}{}}
			props = props["parameters"].(map[string]interface{})
			props["keepalived"] = map[string]interface{}{}
			props["keepalived"].(map[string]interface{})["ooo"] = "kkkoo"
			props["iii"] = "ii"
			props["metadata"] = map[string]interface{}{}
			props["metadata"].(map[string]interface{})["network"] = "testnetwork"
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, props, nil, nil)
		})

		It("returns no error", func() {
			Expect(generateErr).NotTo(HaveOccurred())
		})

		It("returns assigned prop properties", func() {
			Expect(len(manifest.InstanceGroups)).To(BeNumerically(">=", 1))
			Expect(manifest.Properties["keepalived"].(map[string]interface{})["ooo"]).To(Equal("kkkoo"))
			// }
		})
	})

	Context("require parameter missing msg ", func() {
		JustBeforeEach(func() {
			props := map[string]interface{}{"parameters": map[string]interface{}{}}
			props = props["parameters"].(map[string]interface{})
			props["keepalived"] = map[string]interface{}{}
			props["keepalived"].(map[string]interface{})["ooo"] = "kkkoo"
			props["iii"] = "ii"
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, props, nil, nil)
		})

		It("returns error", func() {
			Expect(generateErr).To(HaveOccurred())
		})

		It("contains error msg", func() {
			Expect(generateErr.Error()).To(ContainSubstring("metadata.network"))
			// Expect(generateErr.Error()).To(ContainSubstring("metadata_config.network"))
		})

	})

	Context("dynamic network", func() {
		JustBeforeEach(func() {
			props := map[string]interface{}{"parameters": map[string]interface{}{}}
			props = props["parameters"].(map[string]interface{})
			props["metadata"] = map[string]interface{}{}
			props["metadata"].(map[string]interface{})["network"] = "testnetwork"
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, props, nil, nil)
		})

		It("returns no error", func() {
			Expect(generateErr).NotTo(HaveOccurred())
		})

		It("returns props network", func() {
			Expect(len(manifest.InstanceGroups)).To(BeNumerically(">=", 1))
			for _, ig := range manifest.InstanceGroups {
				Expect(ig.Networks[0].Name).To(Equal("testnetwork"))
			}
		})
	})

	Context("dynamic network -2 ", func() {
		JustBeforeEach(func() {
			props := map[string]interface{}{"parameters": map[string]interface{}{}}
			props = props["parameters"].(map[string]interface{})
			props["metadata"] = map[string]interface{}{}
			props["metadata"].(map[string]interface{})["network"] = []string{"testnetwork", "testnetwork2"}
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, props, nil, nil)
		})

		It("returns no error", func() {
			//Also generate manifest support multiple network, but persistent does not know how to pickup
			//muliple corresponding values for each network yet.
			Expect(generateErr).To(HaveOccurred())
		})

		// It("returns props network", func() {
		// 	fmt.Println(fmt.Sprintf("..... %+v   ", manifest))
		// 	Expect(len(manifest.InstanceGroups)).To(BeNumerically(">=", 1))
		// 	for _, ig := range manifest.InstanceGroups {
		// 		Expect(ig.Networks[0].Name).To(Equal("testnetwork"))
		// 		Expect(ig.Networks[1].Name).To(Equal("testnetwork2"))
		// 	}
		// })
	})

	Context("plan migrations", func() {
		var (
			previousPlan serviceadapter.Plan
		)
		BeforeEach(func() {
			planInstanceGroups = []serviceadapter.InstanceGroup{{Name: "haproxy", Instances: 2}}
			plan = serviceadapter.Plan{InstanceGroups: planInstanceGroups}
		})
		JustBeforeEach(func() {
			previousPlan = serviceadapter.Plan{InstanceGroups: previousPlanInstanceGroups}
			stderr.Reset()
			props := map[string]interface{}{"parameters": map[string]interface{}{}}
			props = props["parameters"].(map[string]interface{})
			props["metadata"] = map[string]interface{}{}
			props["metadata"].(map[string]interface{})["network"] = "testnetwork"
			props["parameters"] = map[string]interface{}{}
			manifest, generateErr = manifestGenerator.GenerateManifest(serviceDeployment, plan, props, nil, &previousPlan)
		})

		Context("when the previous plan had more haproxy", func() {
			BeforeEach(func() {
				previousPlanInstanceGroups = []serviceadapter.InstanceGroup{{Name: "haproxy", Instances: 3}}
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
