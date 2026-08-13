module github.com/axisrobo/ONTOVELA-open/live

go 1.25.5

require (
	github.com/axisrobo/ONTOVELA-open/adapters/effect v0.0.0
	github.com/axisrobo/ONTOVELA-open/adapters/harmovela v0.0.0
	github.com/axisrobo/ONTOVELA-open/adapters/httpwebhook v0.0.0
	github.com/axisrobo/ONTOVELA-open/adapters/mqtt v0.0.0
	github.com/axisrobo/ONTOVELA-open/adapters/opcua v0.0.0
	github.com/axisrobo/ONTOVELA-open/sdk/go v0.0.0
)

replace (
	github.com/axisrobo/ONTOVELA-open/adapters/effect => ../adapters/effect
	github.com/axisrobo/ONTOVELA-open/adapters/harmovela => ../adapters/harmovela
	github.com/axisrobo/ONTOVELA-open/adapters/httpwebhook => ../adapters/httpwebhook
	github.com/axisrobo/ONTOVELA-open/adapters/mqtt => ../adapters/mqtt
	github.com/axisrobo/ONTOVELA-open/adapters/opcua => ../adapters/opcua
	github.com/axisrobo/ONTOVELA-open/sdk/go => ../sdk/go
)
