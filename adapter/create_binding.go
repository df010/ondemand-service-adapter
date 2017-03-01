package adapter

import (
	"errors"
	"fmt"

	"github.com/df010/ondemand-service-adapter/config"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

func (b *Binder) CreateBinding(bindingId string, boshVMs bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) (serviceadapter.Binding, error) {

	config.GetConfigInstance().GetCredentials(boshVMs)
	redisHosts := boshVMs["redis"]
	if len(redisHosts) == 0 {
		b.StderrLogger.Println("no VMs for instance group redis")
		return serviceadapter.Binding{}, errors.New("")
	}

	var redisAddresses []interface{}
	for _, redisHost := range redisHosts {
		redisAddresses = append(redisAddresses, fmt.Sprintf("%s:6379", redisHost))
	}

	return serviceadapter.Binding{
		Credentials: map[string]interface{}{
			"redis": redisAddresses,
		},
	}, nil
}
