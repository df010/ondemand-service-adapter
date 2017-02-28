package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Used struct {
	Values     []interface{} `yaml:"values"`
	Deployment string        `yaml:"deployment"`
}

const (
	VALUE_FORMAT_IP_RANGE     = "ip_range"
	VALUE_FORMAT_NUMBER_RANGE = "number_range"
	VALUE_FORMAT_NUMBER       = "number"
	STORE_FOLDER              = "/var/vcap/store/broker/adapter/"
	STORE_FILE                = STORE_FOLDER + "data.yml"
)

type Input struct {
	Key         string        `yaml:"key"`
	Valueformat string        `yaml:"valueformat"`
	Valuemap    string        `yaml:"valuemap"`
	Value       interface{}   `yaml:"value"`
	Plan        string        `yaml:"plan"`
	Available   []interface{} `yaml:"available"`
	Used        []Used        `yaml:"used"`
}
type Persistent struct {
	Inputs []Input `yaml:"inputs"`
}
type ValueRequest struct {
	Key      string      `yaml:"key"`
	Value    interface{} `yaml:"value"`
	Number   int         `yaml:"number"`
	Specific bool        `yaml:"managed"` //
}
type PersistentRequest struct {
	Plan       string         `yaml:"plan"`
	Deployment string         `yaml:"development"`
	Values     []ValueRequest `yaml:"values"`
}

type PersistentResponse struct {
	Result map[string]interface{} `yaml:"result"`
}

func numberRangeToValues(from string, to string, format string) ([]string, error) {
	ifrom, err1 := strconv.Atoi(from)
	if err1 != nil {
		return nil, err1
	}

	ito, err2 := strconv.Atoi(to)
	if err2 != nil {
		return nil, err2
	}

	result := make([]string, ito-ifrom+1)
	for i := ifrom; i <= ito; i++ {
		result[i-ifrom] = strconv.Itoa(i)
	}
	return result, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func ipRangeToValues(from string, to string, format string) ([]string, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	ifrom := net.ParseIP(from)
	ip := net.ParseIP(from)
	ito := net.ParseIP(to)
	if ifrom.To4() == nil || ito.To4() == nil {
		return nil, fmt.Errorf("either %s or %s are not valid ip address ", from, to)
	}
	length := 0
	result := make([]string, 100)

	for bytes.Compare(ip, ifrom) >= 0 && bytes.Compare(ip, ito) <= 0 {
		if length == len(result) {
			result = append(result, make([]string, 100)...)
		}
		result[length] = ip.To4().String()
		inc(ip)
		length++
	}
	return result[0:length], nil
}

func rangeToValues(from string, to string, format string) ([]string, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if format == VALUE_FORMAT_IP_RANGE {
		return ipRangeToValues(from, to, format)
	} else {
		return numberRangeToValues(from, to, format)
	}
}

func (a *Persistent) usedToMap(input *Input) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	for _, used := range input.Used {
		for _, value := range used.Values {
			result[value] = value
		}
	}
	return result
}

func arrConvert(in []string, typ string) []interface{} {
	out := make([]interface{}, len(in))
	for index, value := range in {
		if typ == VALUE_FORMAT_NUMBER || typ == VALUE_FORMAT_NUMBER_RANGE {
			val, _ := strconv.Atoi(value)
			out[index] = val
		} else {
			out[index] = value
		}
	}
	return out
}

func (a *Persistent) initValue(input *Input, mapping *Input_Mapping) error {
	var vals []interface{}
	input.Valueformat = mapping.Valueformat
	input.Valuemap = mapping.Valuemap
	if reflect.ValueOf(input.Value).Kind() == reflect.String {
		commanSepVals := strings.Split(input.Value.(string), ",")
		var strVals []string
		for i := 0; i < len(commanSepVals); i++ {
			hypenSepVals := strings.Split(commanSepVals[i], "-")
			if len(hypenSepVals) == 1 {
				vals = append(vals, strings.TrimSpace(hypenSepVals[0]))
			} else if len(hypenSepVals) == 2 {
				tvals, _ := rangeToValues(hypenSepVals[0], hypenSepVals[1], input.Valueformat)
				strVals = append(strVals, tvals...)
			} else {
				return fmt.Errorf("unable to parse value: %s ", commanSepVals[i])
			}
		}
		vals = arrConvert(strVals, input.Valueformat)
	} else {
		vals = make([]interface{}, 1)
		vals[0] = input.Value
	}

	availableVals := make([]interface{}, len(vals))
	length := 0
	useds := a.usedToMap(input)
	for n := 0; n < len(vals); n++ {
		if useds[vals[n]] == nil {
			availableVals[length] = vals[n]
			length++
		}
	}
	input.Available = availableVals[0:length]
	return nil
}

func (a *Persistent) getInput(plan string, key string) *Input {
	for i := 0; i < len(a.Inputs); i++ {
		if a.Inputs[i].Key == key && a.Inputs[i].Plan == plan {
			return &a.Inputs[i]
		}
	}
	return nil
}

func (a *Persistent) initFor(request *PersistentRequest) error {

	for n := 0; n < len(request.Values); n++ {
		input := a.getInput(request.Plan, request.Values[n].Key)
		if input == nil {
			input = &Input{Key: request.Values[n].Key, Plan: request.Plan}
			a.Inputs = append(a.Inputs, *input)
			input = &a.Inputs[len(a.Inputs)-1]
		}
		mapping := GetConfigInstance().getInputMapping(request.Values[n].Key)
		if mapping == nil {
			return nil
		}
		if input.Value == request.Values[n].Value && input.Valueformat == mapping.Valueformat && input.Valuemap == mapping.Valuemap {
			return nil //already initiated
		}
		input.Value = request.Values[n].Value
		input.Valueformat = mapping.Valueformat
		input.Valuemap = mapping.Valuemap
		a.initValue(input, mapping)
	}
	// Input {Plan=plan, Key=key};
	return nil
}

func (a *Persistent) init(request *PersistentRequest) error {

	if _, err := os.Stat(STORE_FILE); os.IsNotExist(err) || os.IsPermission(err) {
		content, err := yaml.Marshal(Persistent{})
		if err != nil {
			return err
		}
		err = os.MkdirAll(STORE_FOLDER, 0744)
		err = ioutil.WriteFile(STORE_FILE, content, 0744)
		if err != nil {
			return err
		}
	}

	yamlFile, _ := ioutil.ReadFile(STORE_FILE)
	err := yaml.Unmarshal(yamlFile, a)
	if err != nil {
		return err
	}

	if request != nil {
		a.initFor(request)
	}

	return nil
}

func (a *Persistent) save() error {
	data, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	if len(string(data)) < 50 {
		panic(errors.New("there are no way persistent become empty after init, err"))
	}

	return ioutil.WriteFile(STORE_FILE, data, 0744)
}

func (a *Persistent) findInput(plan string, key string) *Input {
	for i := 0; i < len(a.Inputs); i++ {
		if a.Inputs[i].Key == key && a.Inputs[i].Plan == plan {
			return &a.Inputs[i]
		}
	}
	return nil
}

// func (a *Persistent) findInputs(plan string) []Input {
// 	result := make([]Input, 10)
// 	length := 0
// 	for i := 0; i < len(a.Inputs); i++ {
// 		if a.Inputs[i].Plan == plan {
// 			if length >= len(result) {
// 				result = append(result, make([]Input, 10)...)
// 			}
// 			result[length] = a.Inputs[i]
// 			length++
// 		}
// 	}
// 	return result[0:length]
// }

func (a *Persistent) findUsed(plan string, deployment string, key string, input *Input) *Used {
	if input.Plan != plan || input.Key != key {
		return nil
	}

	for i := 0; i < len(input.Used); i++ {
		if input.Used[i].Deployment == deployment {
			return &input.Used[i]
		}
	}
	return nil
}

func getRequestNumber(valuemap string, number int) int {
	if valuemap == "1" {
		return 1
	}
	maps := strings.Split(valuemap, ":")
	if len(maps) != 2 || maps[1] != "all" {
		return 0
	}
	if maps[0] == "any" {
		return number
	}
	i, _ := strconv.Atoi(maps[0])
	return i
}

func returnUsedValues(valuemap string, used *Used) interface{} {
	if valuemap == "1" || valuemap == "1:all" {
		return used.Values[0]
	} else {
		return used.Values
	}
}
func (a *Persistent) setValues(deployment string, request *ValueRequest, input *Input, used *Used) (interface{}, error) {
	if input.Valuemap == "1" {
		return input.Value, nil
	}
	if used == nil {
		input.Used = append(input.Used, Used{Deployment: deployment})
		used = &input.Used[len(input.Used)-1]
	}
	// if used.Values == nil {
	// 	used.Values = make([]string, 0)
	// }

	number := getRequestNumber(input.Valuemap, request.Number)

	if len(used.Values) == number {
		return returnUsedValues(input.Valuemap, used), nil
	}

	if len(used.Values) > number {
		left := used.Values[0:number]
		del := used.Values[number:]
		used.Values = left
		input.Available = append(input.Available, del...)
	} else {
		if len(input.Available) < number {
			return nil, fmt.Errorf("no enough values to allocate for %v", *request)
		}
		allo := input.Available[0:number]
		input.Available = input.Available[number:]
		used.Values = append(used.Values, allo...)
	}

	if len(used.Values) == 0 {
		a.remove(input, used)
	}

	return returnUsedValues(input.Valuemap, used), nil
}

func (a *Persistent) remove(input *Input, used *Used) {
	i := 0
	length := len(input.Used)
	for ; i < length; i++ {
		if used == &input.Used[i] {
			break
		}
	}
	if i == length-1 {
		input.Used = input.Used[0:length]
	} else if i == 0 {
		input.Used = input.Used[1:length]
	} else {
		input.Used = append(input.Used[0:i], input.Used[i+1:]...)
	}
}

func (a *Persistent) reset() {
	a.Inputs = nil
}

func (a *Persistent) Allocate0(request *PersistentRequest) (map[string]interface{}, error) {
	if len(request.Values) == 0 {
		return nil, nil
	}
	a.reset()
	err := a.init(request)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	for _, rval := range request.Values {
		mapping := GetConfigInstance().getInputMapping(rval.Key)
		if rval.Specific || mapping == nil {
			result[rval.Key] = rval.Value // no config means use what service provide
			continue
		}
		input := a.findInput(request.Plan, rval.Key)
		if input == nil {
			return nil, errors.New("fail to allocate values for request, input not found")
		}
		used := a.findUsed(request.Plan, request.Deployment, rval.Key, input)
		val, err := a.setValues(request.Deployment, &rval, input, used)
		if err != nil {
			return nil, err
		}
		result[rval.Key] = val
	}
	err = a.save()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func combine(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	for key, value := range b {
		a[key] = value
	}
	return a
}

func flatMap(prefix string, val map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range val {
		if reflect.TypeOf(value).Kind() == reflect.Map {
			result = combine(result, flatMap(prefix+key+".", (value.(map[string]interface{}))))
		} else {
			(result)[prefix+key] = value
		}
	}
	return result
}

func getValue(val map[string]interface{}, keys []string) map[string]interface{} {
	data := val
	for i := 0; i < len(keys); i++ {
		// if data[i] == nil {
		// 	data[i] = make(map[string]interface{})
		// }
	}
	return data
}
func compactMap(val map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range val {
		keys := strings.Split(key, ".")
		data := result
		for i := 0; i < len(keys); i++ {
			if i == len(keys)-1 {
				data[keys[i]] = value
			} else {
				tmp := data[keys[i]]
				if tmp == nil {
					data[keys[i]] = make(map[string]interface{})
				}
				data = data[keys[i]].(map[string]interface{})
			}
		}
	}
	return result
}

func (a *Persistent) Allocate(properties map[string]interface{}, plan string, deployment string) (map[string]interface{}, error) {
	requestData := flatMap("", properties)
	request := PersistentRequest{Plan: plan, Deployment: deployment}
	for key, value := range requestData {
		request.Values = append(request.Values, ValueRequest{Key: key, Value: value.(string)})
	}
	result, err := a.Allocate0(&request)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return compactMap(result), nil
	}

	return nil, nil

}

func (a *Persistent) Release(plan string, deployment string) error {
	a.reset()
	err := a.init(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(a.Inputs); i++ {
		for n := 0; n < len(a.Inputs[i].Used); n++ {
			used := &(a.Inputs[i].Used[n])
			if deployment == used.Deployment {
				a.Inputs[i].Used = append(a.Inputs[i].Used[0:n], a.Inputs[i].Used[n+1:]...)
				a.Inputs[i].Available = append(a.Inputs[i].Available, used.Values...)
			}
		}
	}
	a.save()
	return nil
}
