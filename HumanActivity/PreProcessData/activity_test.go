package PreProcessData

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

var activityMetadata *activity.Metadata

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Read sample CSV data.  Results are row-based which emulates the expected input form
func readCSV() [][]float64 {
	var data [][]float64
	dat, err := ioutil.ReadFile("wism_3_activities_one_sample.csv")
	check(err)
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

func getActivityMetadata() *activity.Metadata {

	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}

		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}

	return activityMetadata
}

func TestCreate(t *testing.T) {

	act := NewActivity(getActivityMetadata())

	if act == nil {
		t.Error("Activity Not Created")
		t.Fail()
		return
	}
}

func TestEval(t *testing.T) {

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		t.Failed()
	// 		t.Errorf("panic during execution: %v", r)
	// 	}
	// }()

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	//setup attrs

	data := readCSV()

	tc.SetInput("data", data)
	act.Eval(tc)

	//check result attr
	fmt.Println(tc.GetOutput("predictors"))
}
