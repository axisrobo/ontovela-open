module github.com/axisrobo/ONTOVELA-open/adapters/redis

go 1.25.5

require (
	github.com/axisrobo/ONTOVELA-open/adapters/base v0.0.0
	github.com/axisrobo/ONTOVELA-open/sdk/go v0.0.0
)

replace github.com/axisrobo/ONTOVELA-open/adapters/base => ../base

replace github.com/axisrobo/ONTOVELA-open/sdk/go => ../../sdk/go
