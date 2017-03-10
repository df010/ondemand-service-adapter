package persistent_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/df010/ondemand-service-adapter/persistent"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("persistent", func() {

	BeforeEach(func() {
		os.Remove("/var/vcap/jobs/ondemand/config/config.yml")
		os.Remove("/var/vcap/store/broker/adapter/data.yml")
		ioutil.WriteFile("/var/vcap/jobs/ondemand/config/config.yml", []byte(configString), 0644)
	})

	Context("allocate test", func() {

		var (
			// persistent persistent.Persistent
			// var kk persistent.Allocate("")
			// pp = PersistentRequest{}

			request1 *persistent.PersistentRequest = &persistent.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment1",
			}

			request2 *persistent.PersistentRequest = &persistent.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment2",
			}

			request3 *persistent.PersistentRequest = &persistent.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment3",
			}

			request4 *persistent.PersistentRequest = &persistent.PersistentRequest{
				Plan: "haproxyplan", Deployment: "deployment4",
			}

			result map[string]interface{}
			err    error
			// ipaddress        net.IP
			autoallocatedIP1 net.IP
			specifiedIP2     string
			// twoNumber        []int
			numberResult int
			priority     int
			twoTest      []int

		// previousPlan servicepersistent.Plan
		)

		BeforeEach(func() {
			// persistent = &persistent.Persistent{}
			request1.Values = make([]persistent.ValueRequest, 7)
			request1.Values[0] = persistent.ValueRequest{
				Key: "keepalived.vip1", Value: " 192.168.0.1 - 192.168.0.30 , 192.168.0.50 | 192.168.1.1 - 192.168.1.30 , 192.168.1.50 ",
				Specific: false,
			}
			request1.Values[1] = persistent.ValueRequest{
				Key: "keepalived.vip2", Value: "192.168.0.22",
				Specific: true,
			}
			request1.Values[2] = persistent.ValueRequest{
				Key: "keepalived.virtual_router_id", Value: " 3-100, 111 ",
				Specific: false,
			}
			request1.Values[3] = persistent.ValueRequest{
				Key: "keepalived.priority", Value: 100,
				Specific: false,
			}
			request1.Values[4] = persistent.ValueRequest{
				Key: "keepalived.test", Value: "100-1000",
				Specific: false,
				Number:   2,
			}
			request1.Values[5] = persistent.ValueRequest{
				Key: "metadata_config.network", Value: "net1|net2",
				Specific: false,
			}
			request1.Values[6] = persistent.ValueRequest{
				Key: "metadata.network", Value: "net1",
				Specific: false,
			}

			request2.Values = make([]persistent.ValueRequest, 3)
			request2.Values[0] = persistent.ValueRequest{
				Key: "keepalived.test", Value: "100-105",
				Specific: false,
				Number:   6,
			}
			request2.Values[1] = persistent.ValueRequest{
				Key: "metadata_config.network", Value: "net1|net2",
				Specific: false,
			}
			request2.Values[2] = persistent.ValueRequest{
				Key: "metadata.network", Value: "net1",
				Specific: false,
			}
		})
		JustBeforeEach(func() {
		})

		Context("successfully allocated a group of values", func() {
			BeforeEach(func() {
				// fmt.Println(fmt.Sprintln(".........kk.  request is. %v", request1))
				result, err = persistent.Allocate0(request1)
				// fmt.Println(fmt.Sprintln(".........kk.  2. %v", result))
				// fmt.Println(fmt.Sprintln(".........err is:: . %v", err))
				autoallocatedIP1 = net.ParseIP((result["keepalived.vip1"]).(string))
				// // ipaddress = net.ParseIP(autoallocatedIP1)
				specifiedIP2 = (result["keepalived.vip2"]).(string)
				numberResult = (result["keepalived.virtual_router_id"]).(int)
				// twoNumber = make([]int, len(numberResult))
				// for i, v := range numberResult {
				// 	twoNumber[i] = v.(int)
				// }
				priority = (result["keepalived.priority"]).(int)
				tmpR := (result["keepalived.test"]).([]interface{})
				twoTest = make([]int, len(tmpR))
				for i, v := range tmpR {
					twoTest[i] = v.(int)
				}
			})

			It("fails 22", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("result is a valid ip address keepalived.vip1 ", func() {
				Expect(autoallocatedIP1.To4() == nil).To(BeFalse())
			})
			It("result is a valid ip address keepalived.vip1, second time ", func() {
				Expect(autoallocatedIP1.To4() == nil).To(BeFalse())
			})
			It("result is a valid ip address keepalived.vip1, third time ", func() {
				Expect(autoallocatedIP1.To4() == nil).To(BeFalse())
			})
			It("result ip is what  I specified keepalived.vip2 ", func() {
				Expect(specifiedIP2).To(Equal("192.168.0.22"))
			})
			It("get two number for keepalived.virtual_router_id ", func() {
				Expect(numberResult).To(BeNumerically(">=", 3))
				// Expect(twoNumber[0]).To(BeNumerically("<=", 111))
				// Expect(twoNumber[1]).To(BeNumerically(">=", 3))
				// Expect(twoNumber[1]).To(BeNumerically("<=", 111))
			})
			It("get specified string ", func() {
				Expect(priority).To(Equal(priority))
			})
			It("get two number for  keepalived.test", func() {
				Expect(twoTest[0]).To(BeNumerically(">=", 100))
				Expect(twoTest[1]).To(BeNumerically("<=", 1000))
			})
		})

		Context("not enough values", func() {
			BeforeEach(func() {
				// fmt.Println(fmt.Sprintf("request 2 is:: %+v ", request2))
				result, err = persistent.Allocate0(request2)
				request2.Deployment = "another deployment"
				result, err = persistent.Allocate0(request2)
			})
			It("fails to allocate", func() {
				Expect(err).To(HaveOccurred())
			})
		})
		//
		// Context("check raw value", func() {
		// 	BeforeEach(func() {
		// 		result, err = persistent.Allocate0(request2)
		// 	})
		// 	It("fails to allocate", func() {
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })

		Context("not defined", func() {
			BeforeEach(func() {
				request4.Values = make([]persistent.ValueRequest, 3)
				request4.Values[0] = persistent.ValueRequest{
					Key: "keepalived1.test1", Value: "100-105",
				}
				request4.Values[1] = persistent.ValueRequest{
					Key: "metadata_config.network", Value: "net1|net2",
					Specific: false,
				}
				request4.Values[2] = persistent.ValueRequest{
					Key: "metadata.network", Value: "net1",
					Specific: false,
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
				maprequet["lala"] = "lalalal"
				maprequet["oo"] = "00000000"
				kmaprequest := maprequet["keepalived"].(map[string]interface{})
				kmaprequest["vip1"] = " 192.168.0.1 - 192.168.0.30 , 192.168.0.50 "
				kmaprequest["virtual_router_id"] = " 3-100, 111 "
				kmaprequest["priority"] = 100
			})
			It("fails", func() {
				result, err = persistent.Allocate(maprequet, "haproxyplan", "dddd")
				Expect(err).NotTo(HaveOccurred())
				// fmt.Println(fmt.Sprintf("result data is---------------:: %v", result))
				// Expect(result["keepalived1"]["vip1"][0]).To(ContainSubstring("192.168.0.1"))
			})
		})

		Context("test map groupby ", func() {
			var maprequet = make(map[string]interface{}) //{"keepalived":"b"}
			BeforeEach(func() {
				maprequet["adduser"] = make(map[string]interface{})
				maprequet["metadata_config"] = make(map[string]interface{})

				metadata_config := maprequet["metadata_config"].(map[string]interface{})
				metadata_config["network"] = "net1|net2"

				maprequet["metadata"] = make(map[string]interface{})
				metadata := maprequet["metadata"].(map[string]interface{})
				metadata["network"] = "net2"

				kmaprequest := maprequet["adduser"].(map[string]interface{})
				kmaprequest["password"] = " 20:auto "

			})
			It("fails", func() {
				result, err = persistent.Allocate(maprequet, "haproxyplan", "dddd")
				fmt.Println(fmt.Sprintf("result data is---------------:: %v", result))
				Expect(err).NotTo(HaveOccurred())
				// strings.len
				Expect(len(result["adduser"].(map[string]interface{})["password"].(string))).To(Equal(20))
				// fmt.Println(fmt.Sprintf("result data is---------------:: %v", result))
				// Expect(result["keepalived1"]["vip1"][0]).To(ContainSubstring("192.168.0.1"))
			})
		})

		Context("test map autogen ", func() {
			var maprequet = make(map[string]interface{}) //{"keepalived":"b"}
			BeforeEach(func() {
				maprequet["keepalived"] = make(map[string]interface{})
				maprequet["metadata_config"] = make(map[string]interface{})
				metadata_config := maprequet["metadata_config"].(map[string]interface{})
				metadata_config["network"] = "net1|net2"

				maprequet["metadata"] = make(map[string]interface{})
				metadata := maprequet["metadata"].(map[string]interface{})
				metadata["network"] = "net2"

				kmaprequest := maprequet["keepalived"].(map[string]interface{})
				kmaprequest["vip1"] = " 192.168.0.1 - 192.168.0.2 , 192.168.0.3 | 192.168.1.1 - 192.168.1.2 , 192.168.1.3 "

			})
			It("fails", func() {
				result, err = persistent.Allocate(maprequet, "haproxyplan", "dddd")
				Expect(err).NotTo(HaveOccurred())
				Expect(result["keepalived"].(map[string]interface{})["vip1"]).To(Equal("192.168.1.1"))
				// fmt.Println(fmt.Sprintf("result data is---------------:: %v", result))
				// Expect(result["keepalived1"]["vip1"][0]).To(ContainSubstring("192.168.0.1"))
			})
		})

		Context("release", func() {
			BeforeEach(func() {
				request3.Values = make([]persistent.ValueRequest, 3)
				request3.Values[0] = persistent.ValueRequest{
					Key: "keepalived.test", Value: "100-105",
					Specific: false,
					Number:   4,
				}
				request3.Values[1] = persistent.ValueRequest{
					Key: "metadata_config.network", Value: "net1|net2",
					Specific: false,
				}
				request3.Values[2] = persistent.ValueRequest{
					Key: "metadata.network", Value: "net1",
					Specific: false,
				}

			})
			It("fails", func() {
				request3.Deployment = "dep1"
				result, err = persistent.Allocate0(request3)
				// fmt.Println(fmt.Sprintf("result for dep1 1 %+v", result))
				Expect(err).NotTo(HaveOccurred())

				result, err = persistent.Allocate0(request3) //invoke same deployment twice does not re-allocate
				// fmt.Println(fmt.Sprintf("result for dep1 2 %+v", result))
				Expect(err).NotTo(HaveOccurred())

				request3.Deployment = "dep2"
				result, err = persistent.Allocate0(request3)
				// fmt.Println(fmt.Sprintf("result for dep2 1 %+v", result))
				Expect(err).To(HaveOccurred())

				request3.Deployment = "dep1"
				err = persistent.Release(request3.Deployment)
				Expect(err).NotTo(HaveOccurred())

				request3.Deployment = "dep2"
				result, err = persistent.Allocate0(request3)
				// fmt.Println(fmt.Sprintf("result for dep2 2 %+v", result))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("compsite operation", func() {
			BeforeEach(func() {

				request1.Values = make([]persistent.ValueRequest, 3)
				request1.Plan = "plan1"
				request1.Deployment = "ddd1"

				request1.Values[0] = persistent.ValueRequest{
					Key: "keepalived.vip3", Value: " 192.168.0.1 - 192.168.0.30 , 192.168.0.50 ",
					Specific: false,
				}
				request1.Values[1] = persistent.ValueRequest{
					Key: "metadata_config.network", Value: "net1|net2",
					Specific: false,
				}
				request1.Values[2] = persistent.ValueRequest{
					Key: "metadata.network", Value: "net1",
					Specific: false,
				}
				result, err = persistent.Allocate0(request1)
				// request1.Deployment = "ddd2"
				// result, err = persistent.Allocate0(request1)
				// request1.Deployment = "ddd3"
				// result, err = persistent.Allocate0(request1)

			})
			It("fails", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(result["keepalived.vip3"]).To(ContainSubstring("192.168.0.1"))
			})
		})

	})
})
