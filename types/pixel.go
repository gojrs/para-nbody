package types

import "math"

// ShiftDirection defines our strict asymmetric state routing enum
type ShiftDirection uint8

const (
	ShiftClockwise        ShiftDirection = iota // Matter path: N -> E -> S -> W
	ShiftCounterClockwise                       // Antimatter path: N -> W -> S -> E
	ShiftStatic                                 // Neutral state anchor
)

type Pos struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type Mal struct {
	Tension int64 `json:"tension"` // The Z-axis cross product anchor value
}

// Wiggle evaluates Mal's backpressure based on baseline energy thresholds
func (m *Mal) Wiggle(currentTension int64) int64 {
	const PlanckLimit = 12000
	if currentTension > PlanckLimit {
		return m.Tension / 2 // Dynamic collapse/relaxation threshold
	}
	return m.Tension
}

type SharedState struct {
	X uint64 `json:"x"`
	Y uint64 `json:"y"`
	Z uint64 `json:"z"`
}

type Alice struct{}

// Expand shifts states away from massive density gradients (Inflationary Push)
func (a Alice) Expand(dx, dy, dz int64, gradientForce int64) (int64, int64, int64) {
	if gradientForce > 0 {
		return -dx * gradientForce, -dy * gradientForce, -dz * gradientForce
	}
	return 0, 0, 0
}

type Bob struct{}

// Contract pulls states toward dense gradients (Gravitational Inward Pull)
func (b Bob) Contract(dx, dy, dz int64, gradientForce int64) (int64, int64, int64) {
	if gradientForce > 0 {
		return dx * gradientForce, dy * gradientForce, dz * gradientForce
	}
	return 0, 0, 0
}

type Pixel struct {
	Alice       *Pos        `json:"alice"`
	Bob         Bob         `json:"bob"`
	Mal         Mal         `json:"mal"`
	Destination SharedState `json:"destination"`
}

func NewPixel(startX, startY, startTension int64, destX, destY, destZ uint64) Pixel {
	return Pixel{
		Alice:       &Pos{X: startX, Y: startY},
		Bob:         Bob{},
		Mal:         Mal{Tension: startTension},
		Destination: SharedState{X: destX, Y: destY, Z: destZ},
	}
}

// CalculateTension derives the localized scalar field tension using pure integer hypots
func (p Pixel) CalculateTension() int64 {
	if p.Alice == nil {
		return 0
	}
	// Pythagoras across all coordinates to evaluate the localized spatial knot magnitude
	aliceSq := p.Alice.X*p.Alice.X + p.Alice.Y*p.Alice.Y
	malSq := p.Mal.Tension * p.Mal.Tension

	return int64(math.Sqrt(float64(aliceSq + malSq)))
}

// DetermineChirality extracts the FSM state machine direction on the fly
func (p Pixel) DetermineChirality() ShiftDirection {
	if p.Alice == nil {
		return ShiftStatic
	}
	if p.Alice.X > 0 && p.Mal.Tension < 0 {
		return ShiftClockwise
	}
	if p.Alice.X < 0 && p.Mal.Tension > 0 {
		return ShiftCounterClockwise
	}
	return ShiftStatic
}
