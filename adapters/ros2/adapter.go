// Package ros2 normalizes common ROS 2 robot state messages into ONTOVELA
// state assertions. It is broker-neutral; a ROS 2 client feeds messages in.
package ros2

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidMessage = errors.New("invalid ROS 2 message")

// Pose is an x/y/z position in meters.
type Pose struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// RobotState is a normalized robot status snapshot.
type RobotState struct {
	RobotID     string    `json:"robot_id"`
	Pose        Pose      `json:"pose"`
	Battery     float64   `json:"battery"`
	Status      string    `json:"status"`
	ObservedAt  time.Time `json:"observed_at"`
	Source      string    `json:"source"`
	EvidenceRef string    `json:"evidence_ref"`
}

func (r RobotState) validate() error {
	if r.RobotID == "" || r.Source == "" || r.EvidenceRef == "" {
		return ErrInvalidMessage
	}
	if r.Battery < 0 || r.Battery > 100 {
		return fmt.Errorf("%w: battery out of range", ErrInvalidMessage)
	}
	return nil
}

// ToAssertionInputs maps a robot state into location, battery, and status
// assertion inputs, all observed.
func ToAssertionInputs(robot RobotState) ([]ontovela.StateAssertionInput, error) {
	if err := robot.validate(); err != nil {
		return nil, err
	}
	poseValue, _ := json.Marshal([]float64{robot.Pose.X, robot.Pose.Y, robot.Pose.Z})
	batteryValue, _ := json.Marshal(robot.Battery)
	statusValue, _ := json.Marshal(robot.Status)
	return []ontovela.StateAssertionInput{
		{SubjectID: robot.RobotID, Property: "location", Value: poseValue, StateKind: ontovela.Observed, EventTime: robot.ObservedAt, Source: robot.Source, EvidenceRef: robot.EvidenceRef},
		{SubjectID: robot.RobotID, Property: "battery", Value: batteryValue, StateKind: ontovela.Observed, EventTime: robot.ObservedAt, Source: robot.Source, EvidenceRef: robot.EvidenceRef},
		{SubjectID: robot.RobotID, Property: "status", Value: statusValue, StateKind: ontovela.Observed, EventTime: robot.ObservedAt, Source: robot.Source, EvidenceRef: robot.EvidenceRef},
	}, nil
}
