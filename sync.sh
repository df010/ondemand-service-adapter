go build cmd/service-adapter/main.go
mv main service-adapter
scp service-adapter ubuntu@10.25.50.123:~/  
ssh ubuntu@10.25.50.123 <<'ENDSSH'
scp -i ~/.ssh/id_rsa ~/service-adapter vcap@200.200.2.13:/var/vcap/packages/ondemand-service-adapter/bin/
ENDSSH
#scp service-adapter vcap@200.200.2.13  <<<TzCNyLROc_uf8gzxOmao8BaoFxR4G9QJ 
