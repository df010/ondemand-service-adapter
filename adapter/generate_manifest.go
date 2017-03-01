package adapter

import (
	"errors"
	"fmt"
	"os"
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
	// servicePlan.InstanceGroups[0]
	instanceGroupName := ""
	for _, grp := range servicePlan.InstanceGroups {
		instanceGroupName = instanceGroupName + "-" + grp.Name
	}

	manifestProperties := map[string]interface{}{}
	manifestProperties = merge(manifestProperties, servicePlan.Properties)
	manifestProperties = merge(manifestProperties, requestParams)

	fmt.Fprintf(os.Stderr, "manifest of properties:: before %+v\n", manifestProperties)
	manifestProperties, err = (&persistent.Persistent{}).Allocate(manifestProperties, instanceGroupName, serviceDeployment.DeploymentName)
	fmt.Fprintf(os.Stderr, "manifest of properties:: after %+v\n", manifestProperties)
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
