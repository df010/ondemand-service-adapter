package integration_tests

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("create-binding subcommand", func() {
	var (
		bindingID            string
		boshManifest         string
		boshVMs              string
		bindingRequestParams string

		stdout   bytes.Buffer
		stderr   bytes.Buffer
		exitCode int
	)

	BeforeEach(func() {
		bindingID = ""    // unused
		boshManifest = "" // unused
		boshVMs = `{
			"redis": ["redis1ip", "redis2ip"]
		}`
		bindingRequestParams = `{}`

		stdout = bytes.Buffer{}
		stderr = bytes.Buffer{}
	})

	JustBeforeEach(func() {
		cmd := exec.Command(serviceAdapterBinPath, "create-binding", bindingID, boshVMs, boshManifest, bindingRequestParams)
		runningBin, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, &stdout), io.MultiWriter(GinkgoWriter, &stderr))
		Expect(err).NotTo(HaveOccurred())
		Eventually(runningBin).Should(gexec.Exit())
		exitCode = runningBin.ExitCode()
	})

	Context("when the parameters are valid", func() {
		It("exits with 0", func() {
			Expect(exitCode).To(Equal(0))
		})

		It("prints a binding to stdout", func() {
			var binding serviceadapter.Binding
			Expect(json.Unmarshal(stdout.Bytes(), &binding)).To(Succeed())
			Expect(binding).To(Equal(
				serviceadapter.Binding{
					Credentials: map[string]interface{}{"redis": []interface{}{"redis1ip:6379", "redis2ip:6379"}},
				},
			))
		})
	})

})
