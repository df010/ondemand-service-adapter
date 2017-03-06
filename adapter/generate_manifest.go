package adapter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/df010/ondemand-service-adapter/config"
	"github.com/df010/ondemand-service-adapter/persistent"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

const OnlyStemcellAlias = "only-stemcell"

func defaultDeploymentInstanceGroupsToJobs() map[string][]string {
	result := make(map[string][]string)
	for _, ig := range config.GetConfigInstance().Instance_Groups {
		result[ig.Name] = ig.Templates
	}
	return result
}

func getJobs() []string {
	result := make([]string, len(config.GetConfigInstance().Instance_Groups))
	for i, ig := range config.GetConfigInstance().Instance_Groups {
		result[i] = ig.Name
	}
	return result
}

func GetValueByKey(prop map[string]interface{}, key string) interface{} {
	paths := strings.Split(key, ".")
	node := prop
	// fmt.Fprintf(os.Stderr, "................ check node1 :  %+v", node)
	for j := 0; j < len(paths); j++ {
		// fmt.Fprintf(os.Stderr, "................ check node2 :  %+v", node[paths[j]])
		// fmt.Fprintf(os.Stderr, "................ check node3 :  %+v", paths[j])
		if j == len(paths)-1 {
			if node[paths[j]] != nil {
				return node[paths[j]]
			}
		} else if node[paths[j]] == nil {
			// fmt.Fprintf(os.Stderr, "................  %+v,     %+v,     %+v", paths, node, j)
			return nil
		} else {
			node = node[paths[j]].(map[string]interface{})
		}
	}
	return nil
}

func getErrorMsg(key string, configProperies map[string]interface{}) string {
	configKey := config.GetConfigInstance().GetConfigKeyForRequired(key)
	// fmt.Fprintf(os.Stderr, fmt.Sprintf("+++++++++++++++++++=config key found:: %v   ", configKey))
	if configKey != "" {
		val := GetValueByKey(configProperies, configKey)
		// fmt.Fprintf(os.Stderr, fmt.Sprintf("+++++++++++++++++++=config key value is::: %v   ", configProperies))

		if val != nil {
			return fmt.Sprintf("required key %v are not found in request, please choose from %v", key, val)
		}
	}
	return fmt.Sprintf("required key %v are not found in request", key)
}

func checkRequiredInputs(requestParams serviceadapter.RequestParameters, configProperies map[string]interface{}) error {
	keys := config.GetConfigInstance().GetRequiredKeys()
	errMsg := ""
	for i := 0; i < len(keys); i++ {
		paths := strings.Split(keys[i], ".")
		node := requestParams
		for j := 0; j < len(paths); j++ {
			if j == len(paths)-1 {
				leaf := node[paths[j]]
				if leaf == nil {
					errMsg = errMsg + getErrorMsg(keys[i], configProperies)
				}
				break
			} else {
				if node[paths[j]] == nil {
					errMsg = errMsg + getErrorMsg(keys[i], configProperies)
					break
				}
				node = node[paths[j]].(map[string]interface{})
			}
		}
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

func (a *ManifestGenerator) GenerateManifest(serviceDeployment serviceadapter.ServiceDeployment,
	servicePlan serviceadapter.Plan,
	requestParams serviceadapter.RequestParameters,
	previousManifest *bosh.BoshManifest,
	previousPlan *serviceadapter.Plan,
) (bosh.BoshManifest, error) {
	// a.StderrLogger.Println(fmt.Sprintf("service plan is.... %v ", servicePlan))
	if previousPlan != nil {
		prev := instanceCounts(*previousPlan)
		current := instanceCounts(servicePlan)
		for _, jobName := range getJobs() {
			if prev[jobName] > current[jobName] {
				a.StderrLogger.Println(fmt.Sprintf("cannot migrate to a smaller plan for %s", jobName))
				return bosh.BoshManifest{}, errors.New("")
			}
		}
	}

	var releases []bosh.Release

	for _, serviceRelease := range serviceDeployment.Releases {
		releases = append(releases, bosh.Release{
			Name:    serviceRelease.Name,
			Version: serviceRelease.Version,
		})
	}
	// a.StderrLogger.Println(fmt.Sprintf("service ... 1 "))
	deploymentInstanceGroupsToJobs := defaultDeploymentInstanceGroupsToJobs()
	// a.StderrLogger.Println(fmt.Sprintf("service ... 2 %v ", deploymentInstanceGroupsToJobs))
	// err := checkInstanceGroupsPresent(getJobs(), servicePlan.InstanceGroups)
	//
	// a.StderrLogger.Println(fmt.Sprintf("service ... 3 "))
	// if err != nil {
	// 	a.StderrLogger.Println(err.Error())
	// 	return bosh.BoshManifest{}, errors.New("Contact your operator, service configuration issue occurred")
	// }

	instanceGroups, err := InstanceGroupMapper(servicePlan.InstanceGroups, serviceDeployment.Releases, OnlyStemcellAlias, deploymentInstanceGroupsToJobs)
	manifestProperties := map[string]interface{}{}
	manifestProperties = merge(manifestProperties, servicePlan.Properties)
	if requestParams != nil {
		tmp := requestParams["parameters"] //.(map[string]interface{})
		if tmp != nil {
			requestParams = tmp.(map[string]interface{})
		}
	}

	// fmt.Fprintf(os.Stderr, "config properties:: before %+v\n", manifestProperties)
	// fmt.Fprintf(os.Stderr, "requestParams properties:: before %+v\n", requestParams)
	err = checkRequiredInputs(requestParams, manifestProperties)
	if err != nil {
		return bosh.BoshManifest{}, err
	}
	manifestProperties = merge(manifestProperties, requestParams)

	networks := getNetworks(manifestProperties)
	if networks != nil {
		igs := make([]bosh.InstanceGroup, len(instanceGroups)) //copy a new array of ig to make sure the
		for i := 0; i < len(instanceGroups); i++ {
			igs[i] = instanceGroups[i]
			igs[i].Networks = networks
		}
		instanceGroups = igs
	}

	// a.StderrLogger.Println(fmt.Sprintf("service ... 4 %v ", instanceGroups))
	if err != nil {
		a.StderrLogger.Println(err.Error())
		return bosh.BoshManifest{}, errors.New("")
	}

	for _, ig := range instanceGroups {
		oneInstanceGroup := &ig
		if len(oneInstanceGroup.Networks) != 1 {
			a.StderrLogger.Println(fmt.Sprintf("expected 1 network for %s, got %d", oneInstanceGroup.Name, len(oneInstanceGroup.Networks)))
			return bosh.BoshManifest{}, errors.New("")
		}
	}
	var planId string
	if manifestProperties["plan_id"] == nil {
		for _, grp := range servicePlan.InstanceGroups {
			planId = planId + "-" + grp.Name
		}
	} else {
		planId = manifestProperties["plan_id"].(string)
	}

	manifestProperties, err = persistent.Allocate(manifestProperties, planId, serviceDeployment.DeploymentName)
	// fmt.Fprintf(os.Stderr, "manifest of properties:: after %+v\n", manifestProperties)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "err occurs for generate manifest:: after %+v\n", err)
		return bosh.BoshManifest{}, err
	}
	var updateBlock = bosh.Update{
		Canaries:        1,
		MaxInFlight:     10,
		CanaryWatchTime: "30000-240000",
		UpdateWatchTime: "30000-240000",
		Serial:          boolPointer(false),
	}

	if servicePlan.Update != nil {
		updateBlock = bosh.Update{
			Canaries:        servicePlan.Update.Canaries,
			MaxInFlight:     servicePlan.Update.MaxInFlight,
			CanaryWatchTime: servicePlan.Update.CanaryWatchTime,
			UpdateWatchTime: servicePlan.Update.UpdateWatchTime,
			Serial:          servicePlan.Update.Serial,
		}
	}

	// a.StderrLogger.Println(fmt.Sprintf("instance groups are .... %v ", instanceGroups))
	result := bosh.BoshManifest{
		Name:     serviceDeployment.DeploymentName,
		Releases: releases,
		Stemcells: []bosh.Stemcell{{
			Alias:   OnlyStemcellAlias,
			OS:      serviceDeployment.Stemcell.OS,
			Version: serviceDeployment.Stemcell.Version,
		}},
		InstanceGroups: instanceGroups,
		Properties:     manifestProperties,
		Update:         updateBlock,
	}

	// a.StderrLogger.Println(fmt.Sprintf("bosh deployment is.... %v ", result))
	return result, nil
}

func merge(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	if b == nil {
		return a
	}
	for k, v := range b {
		a[k] = v
	}
	return a
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func instanceCounts(plan serviceadapter.Plan) map[string]int {
	val := map[string]int{}
	for _, instanceGroup := range plan.InstanceGroups {
		val[instanceGroup.Name] = instanceGroup.Instances
	}
	return val
}

func boolPointer(b bool) *bool {
	return &b
}

func getNetworks(props map[string]interface{}) []bosh.Network {
	if props == nil || props["metadata"] == nil {
		return nil
	}
	network := props["metadata"].(map[string]interface{})["network"]

	if network == nil {
		return nil
	}
	var result []bosh.Network
	if reflect.TypeOf(network).Kind() == reflect.String {
		result = make([]bosh.Network, 1)
		result[0] = bosh.Network{Name: network.(string)}
	} else if reflect.TypeOf(network).Kind() == reflect.Slice {
		netArr := network.([]string)
		result = make([]bosh.Network, len(netArr))
		for i, val := range netArr {
			result[i] = bosh.Network{Name: val}
		}
	}
	return result

}

func checkInstanceGroupsPresent(names []string, instanceGroups []serviceadapter.InstanceGroup) error {
	var missingNames []string

	for _, name := range names {
		if !containsInstanceGroup(name, instanceGroups) {
			missingNames = append(missingNames, name)
		}
	}

	if len(missingNames) > 0 {
		return fmt.Errorf("Invalid instance group configuration: expected to find: '%s' in list: '%s'",
			strings.Join(missingNames, ", "),
			strings.Join(getInstanceGroupNames(instanceGroups), ", "))
	}
	return nil
}

func getInstanceGroupNames(instanceGroups []serviceadapter.InstanceGroup) []string {
	var instanceGroupNames []string
	for _, instanceGroup := range instanceGroups {
		instanceGroupNames = append(instanceGroupNames, instanceGroup.Name)
	}
	return instanceGroupNames
}

func containsInstanceGroup(name string, instanceGroups []serviceadapter.InstanceGroup) bool {
	for _, instanceGroup := range instanceGroups {
		if instanceGroup.Name == name {
			return true
		}
	}

	return false
}
