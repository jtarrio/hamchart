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
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"strings"
	"unicode"

	"github.com/jung-kurt/gofpdf"
)

const leftMargin, rightMargin, topMargin, bottomMargin float64 = 0.25, 0.25, 0.25, 0.25

// PdfChart holds the configuration for the chart.
type PdfChart struct {
	Latitude     float64
	Longitude    float64
	Name         string
	WorldMap     *image.Image
	WidthInches  float64
	HeightInches float64
	DotsPerInch  int
	Metric       bool
}

func pt2In(pt float64) float64 {
	return pt / 72.0
}

func textOutXY(pdf *gofpdf.Fpdf, x float64, y float64, w float64, h float64, text string, align string) {
	pdf.SetXY(x+leftMargin, y+topMargin)
	textOut(pdf, w, h, text, align)
}

func textOut(pdf *gofpdf.Fpdf, w float64, h float64, text string, align string) {
	_, lh := pdf.GetFontSize()
	pdf.CellFormat(w, h*lh, text, "", 2, align, false, 0, "")
}

func smallCapsOut(pdf *gofpdf.Fpdf, centerX float64, baselineY float64, text string) {
	var width float64 = 0
	lower := false
	start := 0
	upperSize, _ := pdf.GetFontSize()
	lowerSize := upperSize * .7
	// Measure
	for i, c := range text {
		if lower != unicode.IsLower(c) {
			if start != i {
				segment := text[start:i]
				if lower {
					segment = strings.ToUpper(segment)
				}
				width += pdf.GetStringWidth(segment)
			}
			start = i
			if lower {
				lower = false
				pdf.SetFontSize(upperSize)
			} else {
				lower = true
				pdf.SetFontSize(lowerSize)
			}
		}
	}
	segment := text[start:]
	if lower {
		segment = strings.ToUpper(segment)
	}
	width += pdf.GetStringWidth(segment)

	// Output
	x := leftMargin + centerX - width/2
	start = 0
	for i, c := range text {
		if lower != unicode.IsLower(c) {
			if start != i {
				segment := text[start:i]
				if lower {
					segment = strings.ToUpper(segment)
				}
				pdf.Text(x, baselineY, segment)
				x += pdf.GetStringWidth(segment)
			}
			start = i
			if lower {
				lower = false
				pdf.SetFontSize(upperSize)
			} else {
				lower = true
				pdf.SetFontSize(lowerSize)
			}
		}
	}
	segment = text[start:]
	if lower {
		segment = strings.ToUpper(segment)
	}
	pdf.Text(x, baselineY, segment)
	if lower {
		pdf.SetFontSize(upperSize)
	}
}

func (c PdfChart) makeTitle() string {
	name := fmt.Sprintf("%s, %s", deg2Str(c.Latitude, "N", "S"), deg2Str(c.Longitude, "E", "W"))
	if c.Name != "" {
		name = fmt.Sprintf("%s (%s)", c.Name, name)
	}
	return fmt.Sprintf("centered on %s", name)
}

func deg2Str(degrees float64, posSymbol string, negSymbol string) string {
	symbol := posSymbol
	if degrees < 0 {
		degrees = -degrees
		symbol = negSymbol
	}
	deg := math.Trunc(degrees)
	frac := (degrees - deg) * 60.0
	min := math.Trunc(frac)
	sec := (frac - min) * 60.0

	out := fmt.Sprintf("%.f° ", deg)
	if min > 0 || sec >= 0.1 {
		out += fmt.Sprintf("%.f' ", min)
	}
	if sec >= 0.1 {
		out += fmt.Sprintf("%.1f\" ", sec)
	}
	return out + symbol
}

func (c PdfChart) projectChart(pdf *gofpdf.Fpdf, centerX float64, centerY float64, radius float64) error {
	chartDiameterPx := int(float64(c.DotsPerInch) * 2 * radius)
	chartImage := Project(*c.WorldMap, math.Pi*c.Latitude/180, math.Pi*c.Longitude/180, chartDiameterPx)
	var out bytes.Buffer
	err := png.Encode(&out, chartImage)
	if err != nil {
		return err
	}
	oldWidth := pdf.GetLineWidth()
	options := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true, AllowNegativePosition: true}
	pdf.RegisterImageOptionsReader("chart.png", options, bytes.NewReader(out.Bytes()))
	pdf.SetLineWidth(pt2In(1))
	pdf.ClipCircle(centerX, centerY, radius, true)
	pdf.SetAlpha(0.75, "")
	pdf.ImageOptions("chart.png", centerX-radius, centerY-radius, radius*2, radius*2, false, options, 0, "")
	pdf.SetAlpha(1, "")
	pdf.ClipEnd()
	pdf.SetLineWidth(oldWidth)
	return nil
}

func (c PdfChart) drawChart(pdf *gofpdf.Fpdf, width float64, top float64, bottom float64) error {
	diameter := bottom - top
	if width < diameter {
		diameter = width
	}
	radius := diameter / 2
	centerX := width/2 + leftMargin
	centerY := topMargin + (top+bottom)/2
	chartRadius := radius - pt2In(8)

	err := c.projectChart(pdf, centerX, centerY, chartRadius)
	if err != nil {
		return err
	}

	pdf.SetFont("Helvetica", "", 8)
	for angle := 0; angle < 360; angle += 10 {
		lineOffset := 0.35
		if (angle % 30) == 0 {
			lineOffset = 0.15
		}

		pdf.TransformBegin()
		pdf.TransformRotate(float64(-angle), centerX, centerY)
		pdf.SetLineWidth(pt2In(1))
		pdf.SetDashPattern([]float64{pt2In(2.5), pt2In(3.5)}, 0)
		pdf.SetDrawColor(255, 255, 255)
		pdf.Line(centerX, centerY-lineOffset, centerX, centerY-chartRadius)
		pdf.SetLineWidth(pt2In(0.5))
		pdf.SetDashPattern([]float64{pt2In(2), pt2In(4)}, 0)
		pdf.SetDrawColor(0, 0, 0)
		pdf.Line(centerX, centerY-lineOffset-pt2In(0.25), centerX, centerY-chartRadius)
		text := fmt.Sprintf("%d", angle)
		textWidth := pdf.GetStringWidth(text)
		pdf.Text(centerX-textWidth/2, centerY-(radius-pt2In(6)), text)
		pdf.TransformEnd()
	}

	var distStep, distLimit float64
	var distName string
	distStep, distLimit, distName = 2000, 20000, "Km"
	if !c.Metric {
		distStep, distLimit, distName = 1500, 12427, "mi"
	}
	pdf.SetFontSize(6)
	for distance := distStep; distance < distLimit; distance += distStep {
		radius := distance * chartRadius / distLimit
		pdf.SetLineWidth(pt2In(1))
		pdf.SetDashPattern([]float64{pt2In(1), pt2In(2)}, pt2In(0.25))
		pdf.SetDrawColor(255, 255, 255)
		pdf.Circle(centerX, centerY, radius, "D")
		pdf.SetLineWidth(pt2In(0.5))
		pdf.SetDashPattern([]float64{pt2In(0.5), pt2In(2.5)}, 0)
		pdf.SetDrawColor(0, 0, 0)
		pdf.Circle(centerX, centerY, radius, "D")
		text := fmt.Sprintf("%.f %s", distance, distName)
		textWidth := pdf.GetStringWidth(text)
		pdf.SetDashPattern([]float64{}, 0)
		pdf.SetDrawColor(255, 255, 255)
		pdf.SetLineWidth(pt2In(1))
		pdf.ClipText(centerX-textWidth/2, centerY+radius+pt2In(2), text, true)
		pdf.Rect(centerX-textWidth, centerY+radius-pt2In(5), textWidth*2, pt2In(10), "F")
		pdf.ClipEnd()
		pdf.ClipText(centerX-textWidth/2, centerY-radius+pt2In(2), text, true)
		pdf.Rect(centerX-textWidth, centerY-radius-pt2In(5), textWidth*2, pt2In(10), "F")
		pdf.ClipEnd()
	}
	if pt2In(3) <= (distStep/4)*chartRadius/distLimit {
		pdf.SetLineWidth(pt2In(.5))
		pdf.SetDrawColor(255, 255, 255)
		pdf.Circle(centerX, centerY, pt2In(3), "D")
		pdf.SetLineWidth(pt2In(0.25))
		pdf.SetDrawColor(0, 0, 0)
		pdf.Circle(centerX, centerY, pt2In(3), "D")
	}
	return nil
}

func readFile(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(file)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Generate outputs a PDF to the given writer with the given configuration.
func (c PdfChart) Generate(out io.Writer) error {
	contentWidth, contentHeight := c.WidthInches-(leftMargin+rightMargin), c.HeightInches-(topMargin+bottomMargin)
	ttf, err := readFile("assets/NotoSerif-Regular.ttf")
	if err != nil {
		return err
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "in", Size: gofpdf.SizeType{Wd: c.WidthInches, Ht: c.HeightInches}})
	pdf.AddUTF8FontFromBytes("NotoSerif", "", ttf)
	pdf.SetMargins(leftMargin, topMargin, rightMargin)
	pdf.SetAutoPageBreak(false, bottomMargin)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(pt2In(.5))
	pdf.SetLineCapStyle("butt")
	pdf.AddPage()
	headingUnits := contentWidth / 8
	pdf.SetFont("Times", "", headingUnits*44)
	smallCapsOut(pdf, contentWidth/2, pt2In(headingUnits*53), "Azimuthal Equidistant Chart")
	pdf.SetFont("NotoSerif", "", headingUnits*16)
	pdf.SetXY(0, pt2In(headingUnits*60))
	pdf.WriteAligned(0, pt2In(headingUnits*19), c.makeTitle(), "C")

	err = c.drawChart(pdf, contentWidth, pdf.GetY()+pt2In(headingUnits*19)-topMargin, contentHeight-pt2In(8+4*6))
	if err != nil {
		return err
	}

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetXY(0, topMargin+contentHeight-pt2In(8+4*6))
	pdf.CellFormat(0, pt2In(8), "Bearings shown are relative to true North. Your magnetic declination will vary over time.", "", 2, "C", false, 0, "")
	pdf.SetFontSize(6)
	pdf.CellFormat(0, pt2In(6), "The information shown on this chart may not be complete or accurate.", "", 2, "C", false, 0, "")
	pdf.CellFormat(0, pt2In(6), "Any geographical boundaries and labels shown on this chart are orientative only and are not intended to support or contradict any sovereignty or property claims.", "", 2, "C", false, 0, "")
	pdf.CellFormat(0, pt2In(6), "Image source: https://visibleearth.nasa.gov/images/57752/blue-marble-land-surface-shallow-water-and-shaded-topography", "", 2, "C", false, 0, "")
	pdf.CellFormat(0, pt2In(6), "Copyright 2020 Jacobo Tarrio http://jacobo.tarrio.org This work is licensed under a Creative Commons Attribution 4.0 International License.", "", 2, "C", false, 0, "")

	pdf.Output(out)
	return nil
}
