package PreProcessData

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	metadata *activity.Metadata
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &MyActivity{metadata: metadata}
}

// Metadata implements activity.Activity.Metadata
func (a *MyActivity) Metadata() *activity.Metadata {
	return a.metadata
}

//Convert to column based arrays--simpler to work with
func asCols(data [][]float64) [3][]float64 {

	N := len(data)
	var cols [3][]float64

	for i := 0; i < N; i++ {
		cols[0] = append(cols[0], data[i][0])
		cols[1] = append(cols[1], data[i][1])
		cols[2] = append(cols[2], data[i][2])
	}

	return cols
}

// p-th moment of data given c center and N values
func moment(data []float64, c float64, p float64, N int) float64 {

	sum := 0.0
	for i := 0; i < N; i++ {
		sum += math.Pow(data[i]-c, p)
	}

	return sum / float64(N)
}

// Calculate mean, sd, skewness of data
func moments(data []float64) (mean float64, sd float64, skew float64) {

	N := len(data)
	m1 := moment(data, 0, 1, N)
	m2 := moment(data, m1, 2, N-1)
	m3 := moment(data, m1, 3, N-1)

	return m1, math.Sqrt(m2), m3 / m2 / math.Sqrt(m2)

}

//25th & 75th percentile of data
func percentile(data []float64) (p25 float64, p75 float64) {

	N := float64(len(data))
	loc25 := int(.25 * N)
	loc75 := int(.75 * N)
	safeData := make([]float64, len(data))
	copy(safeData, data)
	sort.Float64s(safeData)

	return safeData[loc25], safeData[loc75]
}

//Pearson correlation between cols x & y
//xbar is mean of x, sdx is std. dev (same for y)
func pearsonCorr(x []float64, y []float64, xbar float64, ybar float64, sdx float64, sdy float64) float64 {

	N := len(x)
	sum := 0.0

	for i := 0; i < N; i++ {
		sum += (x[i] - xbar) * (y[i] - ybar)
	}

	return (sum / float64(N)) / math.Sqrt(sdx*sdx*sdy*sdy)
}

//Calcluate all statistics and place into a map using required feature names
func stats(data [][]float64) map[string]float64 {

	m := make(map[string]float64)
	dataCols := asCols(data)
	meanX, sdX, skewX := moments(dataCols[0])
	meanY, sdY, skewY := moments(dataCols[1])
	meanZ, sdZ, skewZ := moments(dataCols[2])
	q25X, q75X := percentile(dataCols[0])
	q25Y, q75Y := percentile(dataCols[1])
	q25Z, q75Z := percentile(dataCols[2])
	corrxy := pearsonCorr(dataCols[0], dataCols[1], meanX, meanY, sdX, sdY)
	corrxz := pearsonCorr(dataCols[0], dataCols[2], meanX, meanZ, sdX, sdZ)
	corryz := pearsonCorr(dataCols[1], dataCols[2], meanY, meanZ, sdY, sdZ)
	m["x-axis-mean"] = meanX
	m["x-axis-sd"] = sdX
	m["x-axis-skew"] = skewX
	m["y-axis-mean"] = meanY
	m["y-axis-sd"] = sdY
	m["y-axis-skew"] = skewY
	m["z-axis-mean"] = meanZ
	m["z-axis-sd"] = sdZ
	m["z-axis-skew"] = skewZ
	m["x-axis-q25"] = q25X
	m["x-axis-q75"] = q75X
	m["y-axis-q25"] = q25Y
	m["y-axis-q75"] = q75Y
	m["z-axis-q25"] = q25Z
	m["z-axis-q75"] = q75Z
	m["corr-x-y"] = corrxy
	m["corr-x-z"] = corrxz
	m["corr-z-y"] = corryz

	return m
}

func readCSV(fileName string) [][]float64 {
	var data [][]float64
	dat, err := ioutil.ReadFile(fileName)

	r := csv.NewReader(strings.NewReader(string(dat)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var numbers []float64

		for _, elem := range record {
			i, err := strconv.ParseFloat(elem, 64)
			if err == nil {
				numbers = append(numbers, i)
			}
		}
		data = append(data, []float64(numbers))
	}

	return data
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {

	// Get the activity data from the context
	//data := context.GetInput("data").([][]float64)
	file := context.GetInput("file").(string)

	data := readCSV(file)

	m := stats(data)
	outStr := ""

	//Place all output in required format for input to scoring
	i := 0
	for k, v := range m {
		if i > 0 {
			outStr += ","
		}
		outStr += k + ":" + strconv.FormatFloat(v, 'f', -1, 64)
		i++
	}

	context.SetOutput("predictors", outStr)

	return true, nil
}
