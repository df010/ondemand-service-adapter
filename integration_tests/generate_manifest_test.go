package integration_tests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	//"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	//"github.com/pivotal-cf/on-demand-services-sdk/bosh"
)

var _ = Describe("generate-manifest subcommand", func() {
	//var boolPointer = func(b bool) *bool {
	//	return &b
	//}

	const (
		deploymentName = "some-name"
	)

	var (
		plan              string
		previousPlan      string
		requestParams     string
		previousManifest  string
		serviceDeployment string
		stdout            bytes.Buffer
		stderr            bytes.Buffer
		exitCode          int
	)

	BeforeEach(func() {
		stdout = bytes.Buffer{}
		stderr = bytes.Buffer{}

		serviceDeployment = fmt.Sprintf(`{
			"deployment_name": "%s",
			"releases": [{
				"name": "redis",
				"version": "0+dev.1",
				"jobs": ["redis"]
			}],
			"stemcell": {
				"stemcell_os": "Windows",
				"stemcell_version": "3.1"
			}
		}`, deploymentName)
		plan = `{
		  "instance_groups": [
		    {
		      "name": "redis",
		      "vm_type": "small",
		      "persistent_disk_type": "ten",
		      "networks": [
		        "example-network"
		      ],
		      "azs": [
		        "example-az"
		      ],
		      "instances": 1
		    }
		  ],
		  "properties": {}
		}`
		requestParams = `{"parameters": {}}`
		previousManifest = ``
		previousPlan = "null"
	})

	JustBeforeEach(func() {
		cmd := exec.Command(serviceAdapterBinPath, "generate-manifest", serviceDeployment, plan, requestParams, previousManifest, previousPlan)
		runningBin, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, &stdout), io.MultiWriter(GinkgoWriter, &stderr))
		Expect(err).NotTo(HaveOccurred())
		Eventually(runningBin).Should(gexec.Exit())
		exitCode = runningBin.ExitCode()
	})

	Context("when the parameters are valid", func() {
		It("exits with 0", func() {
			Expect(exitCode).To(Equal(0))
		})

		It("prints a manifest to stdout", func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			expectedManifest, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "expected_manifest.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(MatchYAML(expectedManifest))
		})

		Context("plan migrations", func() {
			Context("when migrating to a plan with fewer redis instances", func() {
				BeforeEach(func() {
					plan = `{
					"instance_groups": [
						{
							"name": "redis",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {}
				}`
					previousPlan = `{
					"instance_groups": [
						{
							"name": "redis",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 4
						}
					],
					"properties": {}
				}`
				})

				It("exits with 1", func() {
					Expect(exitCode).To(Equal(1))
				})

				It("logs a message for the operator", func() {
					Expect(stderr.String()).To(ContainSubstring("cannot migrate to a smaller plan"))
				})
			})

			Context("when migrating to a plan with more redis instances", func() {
				BeforeEach(func() {
					plan = `{
					"instance_groups": [
						{
							"name": "redis",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 2
						}
					],
					"properties": {}
				}`
					previousPlan = `{
					"instance_groups": [
						{
							"name": "redis",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {}
				}`
				})

				It("exits with 0", func() {
					Expect(exitCode).To(Equal(0))
				})
			})
		})
	})

	Context("when the service deployment parameter is invalid JSON", func() {
		BeforeEach(func() {
			serviceDeployment = "sadfsdl;fksajflasdf"
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})

	Context("when the service deployment parameter contains too few elements", func() {
		BeforeEach(func() {
			serviceDeployment = `{}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})

	Context("when there are no releases", func() {
		BeforeEach(func() {
			serviceDeployment = fmt.Sprintf(`{
				"deployment_name": "%s",
				"releases": [],
				"stemcell": {
					"stemcell_os": "Windows",
					"stemcell_version": "3.1"
				}
			}`, deploymentName)
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("prints an error message for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("job 'redis' not provided"))
		})

		It("doesn't print an error message for the user", func() {
			Expect(stdout.String()).To(Equal(""))
		})
	})

	Context("when the plan parameter is invalid json", func() {
		BeforeEach(func() {
			plan = "afasfsafasfadsf"
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})

	Context("when the plan contains no instance groups", func() {
		BeforeEach(func() {
			plan = `{
			  "instance_groups": [],
			  "properties": {}
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("outputs an error message for the CLI user to stdout", func() {
			Expect(stdout.String()).To(ContainSubstring("Contact your operator, service configuration issue occurred"))
		})

		It("outputs an error message to the operator to stderr", func() {
			Expect(stderr.String()).To(ContainSubstring("Invalid instance group configuration: expected to find: 'redis' in list: ''"))
		})
	})

	Context("when the plan does not contain a redis instance group", func() {
		BeforeEach(func() {
			plan = `{
			"instance_groups":
				[{
					"name": "tmp",
					"vm_type": "small",
					"networks": [
						"example-network"
					],
					"azs": [
						"example-az"
					],
					"instances": 1
				}]
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("outputs an error message for the CLI user to stdout", func() {
			Expect(stdout.String()).To(ContainSubstring("Contact your operator, service configuration issue occurred"))
		})

		It("outputs an error message to the operator to stderr", func() {
			Expect(stderr.String()).To(ContainSubstring("Invalid instance group configuration: expected to find: 'redis' in list: 'tmp'"))
		})
	})

	Context("release does not contain provided job", func() {
		BeforeEach(func() {
			serviceDeployment = fmt.Sprintf(`{
						"deployment_name": "%s",
						"releases": [{
							"name": "redis",
							"version": "0+dev.1",
							"jobs": ["agshdj"]
						}],
						"stemcell": {
							"stemcell_os": "Windows",
							"stemcell_version": "3.1"
						}
					}`, deploymentName)
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("prints an error message for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("job 'redis' not provided"))
		})

		It("doesn't print an error message for the user", func() {
			Expect(stdout.String()).To(Equal(""))
		})
	})

	Context("when there is no network defined for tha redis job", func() {
		BeforeEach(func() {
			plan = `{
			  "instance_groups": [
			    {
			      "name": "redis",
			      "vm_type": "small",
			      "persistent_disk_type": "ten",
			      "networks": [],
			      "azs": [
			        "example-az"
			      ],
			      "instances": 1
			    }
			  ],
			  "properties": {}
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("prints an error message for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("expected 1 network for redis, got 0"))
		})

		It("doesn't print an error message for the user", func() {
			Expect(stdout.String()).To(Equal(""))
		})

	})
})
