package types

type NBodyConfigMode uint8

const (
	SeedingModeStandard NBodyConfigMode = iota
	SeedingModeChaos
	SeedingModeSplit
	SeedingModeParity
)

func (nbcm NBodyConfigMode) String() (answer string) {
	switch nbcm {
	case SeedingModeStandard:
		answer = "INVARIANT_CORE"
	case SeedingModeChaos:
		answer = "QUANTUM_CHAOS"
	case SeedingModeSplit:
		answer = "PHASE_SPLIT"
	case SeedingModeParity:
		answer = "EVEN_PARITY"
	}

	return answer
}

type EngineMode uint8

const (
	EngineModeClockwise EngineMode = iota
	EngineModeCounterRotating
)

func (em EngineMode) String() (answer string) {
	switch em {
	case EngineModeClockwise:
		answer = "clockwise"
	case EngineModeCounterRotating:
		answer = "counterRotating"
	}

	return answer
}
