package adapter_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"github.com/df010/ondemand-service-adapter/adapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("persistent", func() {
	os.Remove("/var/vcap/jobs/ondemand/config/config.yml")
	os.Remove("/var/vcap/store/broker/adapter/data.yml")
	ioutil.WriteFile("/var/vcap/jobs/ondemand/config/config.yml", []byte(configString), 0644)

	BeforeEach(func() {
	})

	Context("allocate vip", func() {
		var (
			persistent adapter.Persistent
			request1   *adapter.PersistentRequest = &adapter.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment1",
			}

			request2 *adapter.PersistentRequest = &adapter.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment2",
			}

			request3 *adapter.PersistentRequest = &adapter.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment3",
			}

			request4 *adapter.PersistentRequest = &adapter.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment4",
			}

			result map[string]interface{}
			err    error
			// ipaddress        net.IP
			autoallocatedIP1 net.IP
			specifiedIP2     string
			twoNumber        []string
			priority         string
			twoTest          []string

		// previousPlan serviceadapter.Plan
		)

		BeforeEach(func() {
			persistent = adapter.Persistent{}
			request1.Values = make([]adapter.ValueRequest, 5)
			request1.Values[0] = adapter.ValueRequest{
				Key: "keepalived.vip1", Value: " 192.168.0.1 - 192.168.0.30 , 192.168.0.50 ",
				Specific: false,
			}
			request1.Values[1] = adapter.ValueRequest{
				Key: "keepalived.vip2", Value: " 192.168.0.22 ",
				Specific: true,
			}
			request1.Values[2] = adapter.ValueRequest{
				Key: "keepalived.virtual_router_id", Value: " 3-100, 111 ",
				Specific: false,
			}
			request1.Values[3] = adapter.ValueRequest{
				Key: "keepalived.priority", Value: "100",
				Specific: false,
			}
			request1.Values[4] = adapter.ValueRequest{
				Key: "keepalived.test", Value: "100-1000",
				Specific: false,
				Number:   2,
			}
			request2.Values = make([]adapter.ValueRequest, 1)
			request2.Values[0] = adapter.ValueRequest{
				Key: "keepalived.test", Value: "100-105",
				Specific: false,
				Number:   6,
			}
		})
		JustBeforeEach(func() {
		})

		Context("successfully allocated a group of values", func() {
			BeforeEach(func() {
				// fmt.Println(fmt.Sprintln(".........kk.  request is. %v", request1))
				result, err = persistent.Allocate0(request1)
				// fmt.Println(fmt.Sprintln(".........kk.  2. %v", result))
				autoallocatedIP1 = net.ParseIP((result["keepalived.vip1"]).(string))
				// // ipaddress = net.ParseIP(autoallocatedIP1)
				specifiedIP2 = (result["keepalived.vip2"]).(string)
				twoNumber = (result["keepalived.virtual_router_id"]).([]string)
				priority = (result["keepalived.priority"]).(string)
				twoTest = (result["keepalived.test"]).([]string)
			})

			It("fails 22", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("result is a valid ip address keepalived.vip1 ", func() {

				Expect(autoallocatedIP1.To4() == nil).To(BeFalse())
			})
			It("result ip is what  I specified keepalived.vip2 ", func() {
				Expect(specifiedIP2).To(Equal("192.168.0.22"))
			})
			It("get two number for keepalived.virtual_router_id ", func() {
				Expect(strconv.Atoi(twoNumber[0])).To(BeNumerically(">=", 3))
				Expect(strconv.Atoi(twoNumber[0])).To(BeNumerically("<=", 111))
				Expect(strconv.Atoi(twoNumber[1])).To(BeNumerically(">=", 3))
				Expect(strconv.Atoi(twoNumber[1])).To(BeNumerically("<=", 111))
			})
			It("get specified string ", func() {
				Expect(priority).To(Equal(priority))
			})
			It("get two number for  keepalived.test", func() {
				Expect(strconv.Atoi(twoTest[0])).To(BeNumerically(">=", 100))
				Expect(strconv.Atoi(twoTest[1])).To(BeNumerically("<=", 1000))
			})
		})

		Context("not enough values", func() {
			BeforeEach(func() {
				result, err = persistent.Allocate0(request2)
			})
			It("fails to allocate", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("check raw value", func() {

			BeforeEach(func() {
				result, err = persistent.Allocate0(request2)
			})
			It("fails to allocate", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("not defined", func() {
			BeforeEach(func() {
				request4.Values = make([]adapter.ValueRequest, 1)
				request4.Values[0] = adapter.ValueRequest{
					Key: "keepalived1.test1", Value: "100-105",
				}
			})
			It("fails", func() {
				request3.Deployment = "dep2"
				result, err = persistent.Allocate0(request4)
				Expect(err).NotTo(HaveOccurred())
				Expect(result["keepalived1.test1"]).To(Equal("100-105"))
			})
		})

		Context("test map input ", func() {
			var maprequet = make(map[string]interface{}) //{"keepalived":"b"}
			BeforeEach(func() {
				maprequet["keepalived"] = make(map[string]interface{})
				kmaprequest := maprequet["keepalived"].(map[string]interface{})
				kmaprequest["vip1"] = " 192.168.0.1 - 192.168.0.30 , 192.168.0.50 "
				kmaprequest["virtual_router_id"] = " 3-100, 111 "
			})
			It("fails", func() {
				result, err = persistent.Allocate(maprequet, "haproxyplan", "dddd")
				Expect(err).NotTo(HaveOccurred())
				fmt.Println(fmt.Sprintf("result data is:: %v", result))
				// Expect(result["keepalived1"]["vip1"][0]).To(ContainSubstring("100-105"))
			})
		})

		Context("release", func() {
			BeforeEach(func() {
				request3.Values = make([]adapter.ValueRequest, 1)
				request3.Values[0] = adapter.ValueRequest{
					Key: "keepalived.test", Value: "100-105",
					Specific: false,
					Number:   4,
				}

			})
			It("fails", func() {
				request3.Deployment = "dep1"
				result, err = persistent.Allocate0(request3)
				Expect(err).NotTo(HaveOccurred())

				result, err = persistent.Allocate0(request3) //invoke same deployment twice does not re-allocate
				Expect(err).NotTo(HaveOccurred())

				request3.Deployment = "dep2"
				result, err = persistent.Allocate0(request3)
				Expect(err).To(HaveOccurred())

				request3.Deployment = "dep1"
				err = persistent.Release(request3.Plan, request3.Deployment)
				Expect(err).NotTo(HaveOccurred())

				request3.Deployment = "dep2"
				result, err = persistent.Allocate0(request3)
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})
})
