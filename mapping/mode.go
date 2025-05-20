package mapping

import "math"

func ModeExact(inputValue int32, inputInfo ControlInfo, outputInfo ControlInfo) int32 {
	normalized := float64(inputValue-inputInfo.Minimum) / float64(inputInfo.Maximum-inputInfo.Minimum)
	scaledValue := math.Round(normalized*float64(outputInfo.Maximum-outputInfo.Minimum) + float64(outputInfo.Minimum))
	return int32(scaledValue)
}
