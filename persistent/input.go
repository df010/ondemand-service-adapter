package persistent

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/df010/ondemand-service-adapter/config"
)

type Used struct {
	Values     []interface{} `yaml:"values"`
	Deployment string        `yaml:"deployment"`
	Group      string        `yaml:"group"`
}

type Input struct {
	Key         string      `yaml:"key"`
	Valueformat string      `yaml:"valueformat"`
	Valuemap    string      `yaml:"valuemap"`
	Value       interface{} `yaml:"value"`
	Groupby     string      `yaml:"groupby"`
	Plan        string      `yaml:"plan"`
	Available   Available   `yaml:"available"`
	Used        []Used      `yaml:"used"`
}

func (in *Input) initValue(mapping *config.Input_Mapping, value interface{}, groupConfig string) error {
	if in.Value == value && in.Valueformat == mapping.Valueformat && in.Valuemap == mapping.Valuemap {
		return nil //already initiated
	}
	in.Value = value //request.Values[n].Value
	in.Valueformat = mapping.Valueformat
	in.Valuemap = mapping.Valuemap

	var vals []interface{}
	in.Valueformat = mapping.Valueformat
	in.Valuemap = mapping.Valuemap
	if reflect.ValueOf(in.Value).Kind() == reflect.String {
		in.configToAvailable(groupConfig)
	} else {
		vals = make([]interface{}, 1)
		vals[0] = in.Value
		in.filterForAvailable(vals, "")
	}
	// fmt.Fprintf(os.Stderr, "init input :: after %+v\n", in)
	return nil
}

func (in *Input) filterForAvailable(vals []interface{}, group string) {
	availableVals := make([]interface{}, len(vals))
	length := 0
	useds := in.usedToMap(group)
	for n := 0; n < len(vals); n++ {
		if useds[vals[n]] == nil {
			availableVals[length] = vals[n]
			length++
		}
	}
	in.Available.set(availableVals, group)
}

func (in *Input) configToAvailable(groupConfig string) error {
	if groupConfig == "" {
		arr, err := in.stringConfigToArr(in.Value.(string))
		if err != nil {
			return err
		}
		in.filterForAvailable(arr, "")
	} else {

		grps := strings.Split(groupConfig, "|")
		vals := strings.Split(in.Value.(string), "|")
		if len(grps) != len(vals) {
			return fmt.Errorf("group config %+v and value %+v does not match", groupConfig, in.Value)
		}
		for i := 0; i < len(grps); i++ {
			arr, err := in.stringConfigToArr(vals[i])
			if err != nil {
				return err
			}
			in.filterForAvailable(arr, grps[i])
		}
		// fmt.Println(fmt.Sprintln("ini..................... %+v   ", in))
	}
	return nil
}

func (in *Input) stringConfigToArr(value string) ([]interface{}, error) {

	var vals []interface{} // := make([]interface{}, )
	commanSepVals := strings.Split(value, ",")
	var strVals []string
	for i := 0; i < len(commanSepVals); i++ {
		hypenSepVals := strings.Split(commanSepVals[i], "-")
		if len(hypenSepVals) == 1 {
			vals = append(vals, strings.TrimSpace(hypenSepVals[0]))
		} else if len(hypenSepVals) == 2 {
			tvals, _ := rangeToValues(hypenSepVals[0], hypenSepVals[1], in.Valueformat)
			strVals = append(strVals, tvals...)
		} else {
			return nil, fmt.Errorf("unable to parse value: %s ", commanSepVals[i])
		}
	}
	return arrConvert(strVals, in.Valueformat), nil
}

func (in *Input) usedToMap(group string) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	for _, used := range in.Used {
		for _, value := range used.Values {
			if used.Group == group || group == "" {
				result[value] = value
			}
		}
	}
	return result
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (in *Input) setValues(plan string, deployment string, request *ValueRequest, group string) (interface{}, error) {

	if in.Valuemap == "1" {
		return in.Value, nil
	} else if reflect.ValueOf(request.Value).Kind() == reflect.String {
		r, _ := regexp.Compile("^ *([0-9]*) *: *auto *$")
		ret := r.FindStringSubmatch(request.Value.(string))
		if ret != nil || len(ret) == 2 {
			length, _ := strconv.Atoi(ret[1])
			return randSeq(length), nil
		}
	}

	used := in.findUsed(plan, deployment, request.Key)

	if used == nil {
		in.Used = append(in.Used, Used{Deployment: deployment, Group: group})
		used = &in.Used[len(in.Used)-1]
	}
	// if used.Values == nil {
	// 	used.Values = make([]string, 0)
	// }

	number := getRequestNumber(in.Valuemap, request.Number)

	if len(used.Values) == number {
		return returnUsedValues(in.Valuemap, used), nil
	}

	if len(used.Values) > number {
		left := used.Values[0:number]
		del := used.Values[number:]
		used.Values = left
		in.Available.push(del, used.Group)
	} else {

		if !in.Available.avail(number, used.Group) {
			return nil, fmt.Errorf("no enough values to allocate for %v", *request)
		}
		used.Values = in.Available.pop(number, used.Group)
	}

	if len(used.Values) == 0 {
		in.remove(used)
	}

	return returnUsedValues(in.Valuemap, used), nil
}

func (in *Input) remove(used *Used) {
	i := 0
	length := len(in.Used)
	for ; i < length; i++ {
		if used == &in.Used[i] {
			break
		}
	}
	if i == length-1 {
		in.Used = in.Used[0:length]
	} else if i == 0 {
		in.Used = in.Used[1:length]
	} else {
		in.Used = append(in.Used[0:i], in.Used[i+1:]...)
	}
}

func (in *Input) findUsed(plan string, deployment string, key string) *Used {
	if in.Plan != plan || in.Key != key {
		return nil
	}

	for i := 0; i < len(in.Used); i++ {
		if in.Used[i].Deployment == deployment {
			return &in.Used[i]
		}
	}
	return nil
}
