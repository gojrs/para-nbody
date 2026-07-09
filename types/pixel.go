package types

import "math"

type Pos struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

// Magnitude calculates individual field stress vector length
func (p Pos) Magnitude() float64 {
	return math.Sqrt(float64(p.X*p.X + p.Y*p.Y))
}

type SharedState struct {
	X uint64 `json:"x"`
	Y uint64 `json:"y"`
	Z uint64 `json:"z"`
}

// --- V3 MACHINE COMPONENT DOMAINS ---

type Alice struct{}
type Bob struct{} // 🔵 Pure behavioral engine (Calculation of Alice)

// Expand calculates Alice's push AWAY from high-density gradients
func (a Alice) Expand(dx, dy, dz int64, gradientForce int64) (int64, int64, int64) {
	if gradientForce > 0 {
		return -dx * gradientForce, -dy * gradientForce, -dz * gradientForce
	}
	return 0, 0, 0
}

// Contract projects Bob's pull TOWARD high-density gradients
func (b Bob) Contract(dx, dy, dz int64, gradientForce int64) (int64, int64, int64) {
	if gradientForce > 0 {
		return dx * gradientForce, dy * gradientForce, dz * gradientForce
	}
	return 0, 0, 0
}

type Pixel struct {
	Alice       *Pos        `json:"alice"`       // 🔴 Pointer for high-performance zero-allocation mutations
	Bob         Bob         `json:"bob"`         // 🔵 Pure behavioral struct hook
	Destination SharedState `json:"destination"` // Active advection routing vector
}

// NewPixel acts as your safe factory constructor
func NewPixel(startX, startY int64, destX, destY, destZ uint64) Pixel {
	return Pixel{
		Alice:       &Pos{X: startX, Y: startY},
		Bob:         Bob{},
		Destination: SharedState{X: destX, Y: destY, Z: destZ},
	}
}

// CalculateTension derives the localized scalar field tension using Bob as a calculation of Alice
func (p Pixel) CalculateTension() int64 {
	if p.Alice == nil {
		return 0
	}
	aliceSq := p.Alice.X*p.Alice.X + p.Alice.Y*p.Alice.Y
	virtualBobSq := aliceSq // Bob mirrors Alice geometrically on the fly

	return int64(math.Sqrt(float64(aliceSq + virtualBobSq)))
}
