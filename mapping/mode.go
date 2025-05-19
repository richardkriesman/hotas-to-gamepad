package mapping

import "math"

func ModeExact(inputValue int32, inputInfo ControlInfo, outputInfo ControlInfo) int32 {
	// FIXME: need to handle scaling with support for positive/negative values
	// FIXME: needs to support both digital and analog controls
	return (inputValue / int32(math.Abs(float64(inputInfo.Maximum-inputInfo.Minimum)))) * outputInfo.Maximum
}

func ModeShiftFromPositive(inputValue int32, inputInfo ControlInfo, outputInfo ControlInfo) int32 {
	return ModeExact(inputValue-(inputInfo.Maximum/2), inputInfo, outputInfo)
}
