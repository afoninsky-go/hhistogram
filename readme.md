#

### Test with coverage
```
go test -v -coverprofile cover.out ./...
go tool cover -html=cover.out -o cover.html
open cover.html
```