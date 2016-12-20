package integration_tests

import (
	"bytes"
	"io"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("dashboard_url subcommand", func() {
	var (
		stdout   bytes.Buffer
		exitCode int
		plan     = `{
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
			]
		}`
	)

	BeforeEach(func() {
		stdout = bytes.Buffer{}
		cmd := exec.Command(serviceAdapterBinPath, "dashboard-url", "instance-id", plan, "")
		runningBin, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, &stdout), GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(runningBin).Should(gexec.Exit())
		exitCode = runningBin.ExitCode()
	})

	It("should succeed", func() {
		Expect(exitCode).To(BeZero())
	})

	It("generates a arbitrary url", func() {
		Expect(stdout.String()).To(MatchJSON(`{"dashboard_url": ""}`))
	})
})
