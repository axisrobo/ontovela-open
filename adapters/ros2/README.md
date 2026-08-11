# ONTOVELA ROS 2 Adapter

Normalizes common robot state into location, battery, and status assertions.
A ROS 2 client feeds normalized `RobotState` snapshots in.

```go
inputs, err := ros2.ToAssertionInputs(state)
for _, input := range inputs { /* append via core write path */ }
```

All outputs are `observed`; the adapter never fabricates state kinds.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
