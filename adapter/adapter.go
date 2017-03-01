package adapter

import (
	"log"

	"github.com/df010/ondemand-service-adapter/config"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

type ManifestGenerator struct {
	StderrLogger *log.Logger
	config       config.Config
}

type Binder struct {
	StderrLogger *log.Logger
	config       config.Config
}

var InstanceGroupMapper = serviceadapter.GenerateInstanceGroupsWithNoProperties
