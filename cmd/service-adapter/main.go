package main

import (
	"log"
	"os"

	"github.com/df010/ondemand-service-adapter/adapter"
	"github.com/df010/ondemand-service-adapter/persistent"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

func main() {
	stderrLogger := log.New(os.Stderr, "[ondemand] ", log.LstdFlags)
	manifestGenerator := &adapter.ManifestGenerator{
		StderrLogger: stderrLogger,
	}
	binder := &adapter.Binder{
		StderrLogger: stderrLogger,
	}
	// stderrLogger.Println(fmt.Sprintf("logllll ...  %v ", os.Args))
	if os.Args[1] == "release" {
		err := persistent.ReleaseOthers(os.Args[2])
		if err != nil {
			stderrLogger.Fatal(err)
		}
	} else {
		serviceadapter.HandleCommandLineInvocation(os.Args, manifestGenerator, binder, &adapter.DashboardUrlGenerator{})
	}
}
