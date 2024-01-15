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
	"hamchart/assets"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type chartHandler struct {
	worldMap image.Image
}

func NewChartHandler() (http.Handler, error) {
	photo := bytes.NewReader(assets.EarthPhoto)
	worldMap, err := png.Decode(photo)
	if err != nil {
		return nil, err
	}
	return &chartHandler{worldMap: worldMap}, nil
}

func getNum(req *http.Request, name string, def float64) float64 {
	vals := req.PostForm[name]
	if len(vals) == 0 {
		return def
	}
	ret, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		return def
	}
	return ret
}

func getStr(req *http.Request, name string, def string) string {
	vals := req.PostForm[name]
	if len(vals) == 0 {
		return def
	}
	return vals[0]
}

func getBool(req *http.Request, name string) bool {
	vals := req.PostForm[name]
	return len(vals) > 0
}

func getSize(req *http.Request, name string, def string) (width float64, height float64) {
	size := strings.ToLower(getStr(req, name, def))
	switch size {
	case "letter":
		return 8.5, 11
	case "a3":
		return 297.0 / 25.4, 420.0 / 25.4
	default:
		return 210.0 / 25.4, 297.0 / 25.4
	}
}

func (h *chartHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Printf("Error parsing request parameters: %s", err)
		http.Error(w, "Error parsing request parameters", 400)
		return
	}

	latitude := getNum(req, "latitude", 0)
	longitude := getNum(req, "longitude", 0)
	name := strings.TrimSpace(getStr(req, "name", ""))
	metric := getBool(req, "metric")
	width, height := getSize(req, "size", "a4")
	chart := PdfChart{
		Latitude:     latitude,
		Longitude:    longitude,
		Name:         name,
		WidthInches:  width,
		HeightInches: height,
		DotsPerInch:  300,
		Metric:       metric,
		WorldMap:     h.worldMap,
	}

	log.Printf("Generating chart for (%f,%f)[%s] size:%fx%f metric:%t", latitude, longitude, name, width, height, metric)
	w.Header()["Content-Type"] = []string{"application/pdf"}
	chart.Generate(w)
	if err != nil {
		http.Error(w, fmt.Sprintf("An error occurred generating the chart: %s", err), 500)
		log.Printf("Error while generating the chart: %s", err)
	}
}
