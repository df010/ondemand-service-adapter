# redis-service-adapter

This project relies on the Pivotal Services SDK, available to customers and partners for download at [network.pivotal.io](http://network.pivotal.io)

---

An example of an [on-demand broker](http://docs.pivotal.io/on-demand-service-broker) service adapter for Kafka.

[Example BOSH release](https://github.com/df010/redis-service-adapter-release) for this service adapter.

---

## Development

1. If you haven't already arrived in this repository as a submodule of [its bosh release](https://github.com/df010/redis-service-adapter-release), then `go get github.com/df010/redis-service-adapter`
1. Install [Ginkgo](https://onsi.github.io/ginkgo/) if you haven't already: `go get github.com/onsi/ginkgo/ginkgo`
1. `cd $GOPATH/src/github.com/df010/redis-service-adapter`
1. `./scripts/run-tests.sh`

