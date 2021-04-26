#

### Test with coverage
```
go test -v -coverprofile cover.out ./...
go tool cover -html=cover.out -o cover.html
open cover.html
```

### Integration tests
```
aws s3 cp s3://mambu-staging-alb-logs-eu-west-2/debug/big2.log.gz test.log.gz
go run main.go
vector --config-yaml test/vector.yaml
curl -v -X POST --data-binary @./test.log.gz -H "Content-encoding: gzip" http://localhost:7999
```