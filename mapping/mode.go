package mapping

import "math"

var modeMap = map[string]ModeFunction{
	"linear": ModeLinear,
}

func GetMode(name string) ModeFunction {
	return modeMap[name]
}

func ModeLinear(inputValue int32, inputInfo ControlInfo, outputInfo ControlInfo) int32 {
	normalized := float64(inputValue-inputInfo.Minimum) / float64(inputInfo.Maximum-inputInfo.Minimum)
	scaledValue := math.Round(normalized*float64(outputInfo.Maximum-outputInfo.Minimum) + float64(outputInfo.Minimum))
	return int32(scaledValue)
}
