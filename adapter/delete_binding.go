package adapter

import (
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

func (b *Binder) DeleteBinding(bindingId string, boshVMs bosh.BoshVMs, manifest bosh.BoshManifest, requestParameters serviceadapter.RequestParameters) error {
	// redisServers := boshVMs["redis"]
	// if len(redisServers) == 0 {
	// 	b.StderrLogger.Println("no VMs for job redis")
	// 	return errors.New("")
	// }
	return nil
}
