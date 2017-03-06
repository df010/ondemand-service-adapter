package adapter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/df010/ondemand-service-adapter/config"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

func (b *Binder) CreateBinding(bindingId string, boshVMs bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) (serviceadapter.Binding, error) {
	credentials, err := getCredentials("haproxy", manifest.Properties, boshVMs)
	if err != nil {
		return serviceadapter.Binding{}, err
	}

	return serviceadapter.Binding{
		Credentials: credentials,
	}, nil
}

func getCredentials(plan string, prop map[string]interface{}, boshVMs bosh.BoshVMs) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, credential := range config.GetConfigInstance().Binding_Credentials {
		if credential.Plan == "" || credential.Plan == plan {
			r, _ := regexp.Compile("(.*)\\[([^\\[\\]]*)\\]\\.*(.*)")
			matches := r.FindStringSubmatch(credential.Value)
			if len(matches) < 4 {
				result[credential.Name] = credential.Value
			} else {
				src := matches[1]
				key := matches[2]
				subkey := matches[3]
				var err error
				if src == "properties" {
					val := GetValueByKey(prop, key)
					if val != nil {
						result[credential.Name] = val
					} else {
						err = fmt.Errorf("%v is not found ", key)
					}
				} else if src == "job" && subkey == "ip" { //currently, ip is the only supported job vm property
					err = putJobProp(credential.Datatype, &result, key, boshVMs, credential.Name)
				}
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return result, nil
}

func putJobProp(typ string, result *map[string]interface{}, key string, boshVMs bosh.BoshVMs, resultKey string) error {
	if typ == "string" {
		return putJobPropString(result, key, boshVMs, resultKey)
	} else {
		return putJobPropArray(result, key, boshVMs, resultKey)
	}

}

func putJobPropString(result *map[string]interface{}, key string, boshVMs bosh.BoshVMs, resultKey string) error {
	empty := true
	for k, v := range transform(getVMPropsMap(key, boshVMs)) {
		empty = false
		if resultKey != "" {
			if (*result)[resultKey] == nil {
				(*result)[resultKey] = ""
			}
			(*result)[resultKey] = (*result)[resultKey].(string) + "," + v
		} else {
			(*result)[k] = v
		}
	}
	if empty {
		return fmt.Errorf("%v is not found ", key)
	} else {
		return nil
	}
}

func putJobPropArray(result *map[string]interface{}, key string, boshVMs bosh.BoshVMs, resultKey string) error {
	empty := true
	for k, v := range getVMPropsMap(key, boshVMs) {
		empty = false
		if resultKey != "" {
			rel := (*result)[resultKey]
			if rel != nil {
				(*result)[resultKey] = append(rel.([]string), v...)
			} else {
				(*result)[resultKey] = v
			}
		} else {
			(*result)[k] = v
		}
	}
	if empty {
		return fmt.Errorf("%v is not found ", key)
	} else {
		return nil
	}

}

func transform(mp map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, val := range mp {
		result[key] = strings.Join(val, ",")
	}
	return result
}

func getVMProps(jobName string, boshVMs bosh.BoshVMs) []string {
	var result []string
	if jobName == "*" || jobName == "" {
		for _, val := range boshVMs {
			result = append(result, val...)
		}
	} else {
		result = append(result, boshVMs[jobName]...)
	}
	return result
}

func getVMPropsMap(jobName string, boshVMs bosh.BoshVMs) map[string][]string {
	result := make(map[string][]string)
	if jobName == "*" || jobName == "" {
		for key, val := range boshVMs {
			result[key] = val
		}
	} else {
		result[jobName] = boshVMs[jobName]
	}
	return result
}
