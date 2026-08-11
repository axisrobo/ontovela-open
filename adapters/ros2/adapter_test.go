package ros2

import (
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func TestToAssertionInputsMapsRobotState(t *testing.T) {
	robot := RobotState{
		RobotID: "robot:WH-17", Pose: Pose{X: 1, Y: 2, Z: 0}, Battery: 82, Status: "ready",
		ObservedAt: time.Now().UTC(), Source: "kinetovela:robot-1", EvidenceRef: "ros/pose-1",
	}
	inputs, err := ToAssertionInputs(robot)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 3 {
		t.Fatalf("inputs = %d", len(inputs))
	}
	properties := map[string]bool{}
	for _, input := range inputs {
		if input.StateKind != ontovela.Observed {
			t.Fatalf("state kind = %q", input.StateKind)
		}
		properties[input.Property] = true
	}
	for _, property := range []string{"location", "battery", "status"} {
		if !properties[property] {
			t.Fatalf("missing property %q", property)
		}
	}
}

func TestToAssertionInputsRejectsInvalidBattery(t *testing.T) {
	robot := RobotState{RobotID: "robot:WH-17", Battery: 120, ObservedAt: time.Now().UTC(), Source: "s", EvidenceRef: "e"}
	if _, err := ToAssertionInputs(robot); err == nil {
		t.Fatal("invalid battery must be rejected")
	}
}
