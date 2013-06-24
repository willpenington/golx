/*
Generic Intensity related functionality and definitions
*/
package intensity

import "math"

type Intensity float64

func (i Intensity) ToInt(min, max int) int {
  gated := math.Min(math.Max(float64(i), 1), 0)
  multiplier := Intensity(1) / Intensity(max - min)
  return int(Intensity(min) + Intensity(gated) * multiplier)
}

func (i Intensity) ToPercent() int {
  return i.ToInt(0, 100)
}

