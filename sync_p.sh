go build cmd/service-adapter/main.go
mv main service-adapter
scp service-adapter vcap@30.16.1.30:/var/vcap/packages/ondemand-service-adapter/bin/
#scp service-adapter vcap@200.200.2.13  <<<TzCNyLROc_uf8gzxOmao8BaoFxR4G9QJ 
