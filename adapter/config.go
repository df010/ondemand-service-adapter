package adapter

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Binding_Credentials []struct {
		Name     string `yaml:"name"`
		Datatype string `yaml:"datatype"`
		Value    string `yaml:"value"`
	}
	Instance_Groups []struct {
		Name      string   `yaml:"name"`
		Templates []string `yaml:"templates"`
	}
}

var config *Config = &Config{}
var once sync.Once

func GetConfigInstance() *Config {
	once.Do(func() {
		yamlFile, _ := ioutil.ReadFile("/var/vcap/sys/run/ondemand-service-adapter/config/config.yml")
		err := yaml.Unmarshal(yamlFile, config)

		if err != nil {
			panic(err)
		}
	})
	return config
}

func (a *Config) getCredentials(boshVMs bosh.BoshVMs) map[string]interface{} {
	result := make(map[string]interface{})
	for _, credential := range a.Binding_Credentials {
		re := regexp.MustCompile("\\[([^\\[\\]]*)\\]")
		jobIpExpr := regexp.MustCompile("JOB\\.(.*)\\.ip")
		var resultCredential interface{}
		jobName := ""
		for _, str := range re.FindAllString(credential.Value, -1) {
			innerStr := strings.TrimPrefix(str, "[")
			innerStr = strings.TrimSuffix(innerStr, "]")
			if jobIpExpr.MatchString(innerStr) {
				jobName = jobIpExpr.FindStringSubmatch(innerStr)[1]
			}
		}
		if jobName != "" {
			if credential.Datatype == "array" {
				arr := make([]string, len(boshVMs[jobName]))
				for i, ip := range boshVMs[jobName] {
					arr[i] = strings.Replace(credential.Value, "[JOB."+jobName+".ip]", ip, -1)
				}
				resultCredential = arr
			}
		}
		if resultCredential != nil {
			result[credential.Name] = resultCredential
		}

	}
	fmt.Println(fmt.Sprintf("result.... %v ", result))
	return result
}
