package adapter

import (
	"log"

	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

type ManifestGenerator struct {
	StderrLogger *log.Logger
	config       Config
}

type Binder struct {
	StderrLogger *log.Logger
	config       Config
}

var InstanceGroupMapper = serviceadapter.GenerateInstanceGroupsWithNoProperties
