# Elvis framework

# Public

```
go mod tidy &&
gofmt -w . &&
git update &&
git tag v0.0.84 &&
git tags
git push origin --tags

go build ./cmd/create-go
gofmt -w . && go run ./cmd/create-go
gofmt -w . && go run github.com/cgalvisleon/elvis/cmd/create-go create
go run github.com/cgalvisleon/elvis/cmd/create-go create

go get -u github.com/cgalvisleon/elvis@v0.0.64
```