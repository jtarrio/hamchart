/*
   Copyright 2020 Jacobo Tarrío Barreiro (http://jacobo.tarrio.org)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package chartgen

import (
	"image"
	"math"
)

// Project creates an azimuthal equidistant projection of the source map, centered on a given latitude and longitude.
func Project(source image.Image, latitude float64, longitude float64, chartWidth int) image.Image {
	output := image.NewRGBA(image.Rect(0, 0, chartWidth, chartWidth))
	bounds := source.Bounds()
	width := float64(bounds.Max.X - bounds.Min.X)
	height := float64(bounds.Max.Y - bounds.Min.Y)
	center := float64(chartWidth) / 2
	centerInt := int(center + 0.5)
	pixelRadians := math.Pi / center
	maxDistance := (center + math.Sqrt(0.5)) * pixelRadians

	latitudeSin := math.Sin(latitude)
	latitudeCos := math.Cos(latitude)

	for py := 0; py < centerInt; py++ {
		y := (center - float64(py) - 0.5) * pixelRadians
		for px := 0; px <= py; px++ {
			x := (float64(px) + 0.5 - center) * pixelRadians
			dist := math.Sqrt(x*x + y*y)
			if dist == 0 {
				ox, oy := latLongToXY(latitude, longitude, width, height)
				output.Set(px, py, source.At(ox, oy))
			} else if dist < maxDistance {
				distSin := math.Sin(dist)
				distCos := math.Cos(dist)
				latPart1 := distCos * latitudeSin
				latPart2 := distSin * latitudeCos / dist
				longPart1 := dist * latitudeCos * distCos
				longPart2 := latitudeSin * distSin
				for octant := range [8]int{} {
					xCoord := x
					yCoord := y
					xOut := px
					yOut := py
					if octant == 1 || octant == 2 || octant == 5 || octant == 6 {
						xCoord, yCoord = yCoord, xCoord
						xOut, yOut = yOut, xOut
					}
					if octant == 1 || octant == 3 || octant == 4 || octant == 6 {
						xCoord = -xCoord
					}
					if octant == 1 || octant == 2 || octant == 4 || octant == 7 {
						yCoord = -yCoord
					}
					if octant >= 2 && octant < 6 {
						xOut = chartWidth - xOut - 1
					}
					if octant >= 4 {
						yOut = chartWidth - yOut - 1
					}
					lat := math.Asin(latPart1 + yCoord*latPart2)
					long := longitude + math.Atan2(xCoord*distSin, longPart1-yCoord*longPart2)
					ox, oy := latLongToXY(lat, long, width, height)
					output.Set(xOut, yOut, source.At(ox, oy))
				}
			}
		}
	}
	return output
}

func latLongToXY(latitude float64, longitude float64, width float64, height float64) (int, int) {
	return int(clamp(1+longitude/math.Pi, 2) * width / 2), int(clamp(1-2*latitude/math.Pi, 2) * height / 2)
}

func clamp(value float64, limit float64) float64 {
	for value < 0 {
		value += limit
	}
	for value >= limit {
		value -= limit
	}
	return value
}
