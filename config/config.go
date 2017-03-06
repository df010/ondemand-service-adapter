package config

import (
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"

	yaml "gopkg.in/yaml.v2"
)

type Input_Mapping struct {
	Key         string `yaml:"key"`
	Valueformat string `yaml:"valueformat"`
	Valuemap    string `yaml:"valuemap"`
	Groupby     string `yaml:"groupby"`
	Required    bool   `yaml:"required"`
}

type Config struct {
	Binding_Credentials []struct {
		Name     string `yaml:"name"`
		Datatype string `yaml:"datatype"`
		Value    string `yaml:"value"`
		Plan     string `yaml:"plan"`
	}
	Instance_Groups []struct {
		Name      string   `yaml:"name"`
		Templates []string `yaml:"templates"`
	}
	Input_Mappings []Input_Mapping
}

var config *Config = &Config{}
var once sync.Once

func GetConfigInstance() *Config {
	once.Do(func() {
		yamlFile, _ := ioutil.ReadFile("/var/vcap/jobs/ondemand/config/config.yml")
		err := yaml.Unmarshal(yamlFile, config)

		if err != nil {
			panic(err)
		}
	})
	return config
}

func (a *Config) GetInputMapping(key string) *Input_Mapping {
	for _, inputMap := range a.Input_Mappings {
		if inputMap.Key == key {
			return &inputMap
		}
	}
	return nil
}

func (a *Config) GetConfigKeyForRequired(key string) string {

	if strings.HasPrefix(key, "metadata.") {
		splits := strings.Split(key, "metadata.")
		return "metadata_config." + splits[1]
	}
	return ""
}

func (a *Config) GetRequiredKeys() []string {
	var result []string
	for i := 0; i < len(a.Input_Mappings); i++ {
		if a.Input_Mappings[i].Required {
			result = append(result, a.Input_Mappings[i].Key)
		} else if a.Input_Mappings[i].Groupby != "" {
			result = append(result, "metadata."+a.Input_Mappings[i].Groupby)
		}
	}
	return result
}

func (a *Config) GetCredentials(boshVMs bosh.BoshVMs) map[string]interface{} {
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
	return result
}
