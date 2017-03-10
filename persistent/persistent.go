package persistent

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/df010/ondemand-service-adapter/config"
	"github.com/nightlyone/lockfile"
	yaml "gopkg.in/yaml.v2"
)

const (
	VALUE_FORMAT_IP_RANGE     = "ip_range"
	VALUE_FORMAT_NUMBER_RANGE = "number_range"
	VALUE_FORMAT_NUMBER       = "number"
	STORE_FOLDER              = "/var/vcap/store/broker/adapter/"
	STORE_FILE                = STORE_FOLDER + "data.yml"
)

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

func (r *PersistentRequest) getGroup(groupby string) string {
	for i := 0; i < len(r.Values); i++ {
		if r.Values[i].Key == "metadata."+groupby {
			return r.Values[i].Value.(string)
		}
	}
	return ""
}

func (r *PersistentRequest) getGroupConfig(groupby string) string {
	for i := 0; i < len(r.Values); i++ {
		if r.Values[i].Key == "metadata_config."+groupby {
			return r.Values[i].Value.(string)
		}
	}
	return ""
}

type Persist struct {
	Inputs []Input `yaml:"inputs"`
}

func (a *Persist) getInput(plan string, key string) *Input {
	for i := 0; i < len(a.Inputs); i++ {
		if a.Inputs[i].Key == key && a.Inputs[i].Plan == plan {
			return &a.Inputs[i]
		}
	}
	return nil
}

func (a *Persist) initFor(request *PersistentRequest) error {

	for n := 0; n < len(request.Values); n++ {

		// fmt.Fprintf(os.Stderr, "try to init for  ::  %+v\n", request.Values[n])
		mapping := config.GetConfigInstance().GetInputMapping(request.Values[n].Key)
		if mapping == nil {
			// fmt.Fprintf(os.Stderr, "no config for the request, exit  %+v\n", 1)
			continue
		}
		input := a.getInput(request.Plan, request.Values[n].Key)
		// fmt.Fprintf(os.Stderr, "init input :: before %+v\n", input)
		if input == nil {
			input = &Input{Key: request.Values[n].Key, Plan: request.Plan, Groupby: mapping.Groupby}
			a.Inputs = append(a.Inputs, *input)
			input = &a.Inputs[len(a.Inputs)-1]
		}
		input.initValue(mapping, request.Values[n].Value, request.getGroupConfig(input.Groupby))
	}
	return nil
}

func (a *Persist) init(request *PersistentRequest) error {

	if _, err := os.Stat(STORE_FILE); os.IsNotExist(err) || os.IsPermission(err) {
		// fmt.Fprintf(os.Stderr, "try to init for ::  %+v\n", STORE_FILE)
		content, err := yaml.Marshal(Persist{})
		if err != nil {
			return err
		}
		err = os.MkdirAll(STORE_FOLDER, 0744)
		// fmt.Fprintf(os.Stderr, "err for make folder ::  %+v\n", err)
		err = ioutil.WriteFile(STORE_FILE, content, 0744)
		// fmt.Fprintf(os.Stderr, "err for write file ::  %+v\n", err)
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
	// fmt.Fprintf(os.Stderr, "after file init is ::  %+v\n", a)
	return nil
}

func (a *Persist) save() error {
	// fmt.Println(fmt.Sprintf("..before amrsh.. %+v", a))
	data, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	// fmt.Println(fmt.Sprintf(".... %+v", string(data)))

	// if len(string(data)) < 50 {
	// 	panic(errors.New("there are no way persistent become empty after init, err"))
	// }

	return ioutil.WriteFile(STORE_FILE, data, 0744)
}

func (a *Persist) findInput(plan string, key string) *Input {
	for i := 0; i < len(a.Inputs); i++ {
		if a.Inputs[i].Key == key && a.Inputs[i].Plan == plan {
			return &a.Inputs[i]
		}
	}
	return nil
}

func (a *Persist) reset() {
	a.Inputs = nil
}

func trylock() lockfile.Lockfile {
	lock, err := lockfile.New(filepath.Join(os.TempDir(), "lock.me.now.lck"))
	if err != nil {
		// fmt.Printf("Cannot init lock. reason: %v", err)
		panic(err) // handle properly please!
	}

	err = lock.TryLock()
	count := 0
	for err != nil && count < 100 {
		// Error handling is essential, as we only try to get the lock.

		time.Sleep(1)
		count++
		err = lock.TryLock()
	}
	if err != nil {
		// fmt.Printf("Cannot lock %q, reason: %v", lock, err)
		panic(err) // handle properly please!
	}
	return lock
}

func Allocate0(request *PersistentRequest) (map[string]interface{}, error) {
	lock := trylock()
	defer lock.Unlock()

	if len(request.Values) == 0 {
		return nil, nil
	}

	a := &Persist{}

	a.reset()
	err := a.init(request)
	if err != nil {
		return nil, err
	}

	// fmt.Println(fmt.Sprintf("result for ----   %+v", a))

	result := make(map[string]interface{})

	for _, rval := range request.Values {
		// fmt.Fprintf(os.Stderr, "try to set value for %+v\n", rval)
		mapping := config.GetConfigInstance().GetInputMapping(rval.Key)
		if rval.Specific || mapping == nil {
			result[rval.Key] = rval.Value // no config means use what service provide
			continue
		}
		input := a.findInput(request.Plan, rval.Key)
		if input == nil {
			return nil, errors.New("fail to allocate values for request, input not found")
		}
		// fmt.Fprintf(os.Stderr, "input :: found for request %+v\n", input)
		val, err := input.setValues(request.Plan, request.Deployment, &rval, request.getGroup(input.Groupby))
		if err != nil {
			return nil, err
		}
		result[rval.Key] = val
	}
	// fmt.Println(fmt.Sprintf("after allocate0  for ----   %+v", a))
	// fmt.Fprintf(os.Stderr, "allocate properties end %+v\n", a)
	err = a.save()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Allocate(properties map[string]interface{}, plan string, deployment string) (map[string]interface{}, error) {
	requestData := flatMap("", properties)
	request := PersistentRequest{Plan: plan, Deployment: deployment}
	for key, value := range requestData {
		request.Values = append(request.Values, ValueRequest{Key: key, Value: value})
	}
	// fmt.Fprintf(os.Stderr, "request is::  %+v\n", request)
	result, err := Allocate0(&request)
	// fmt.Fprintf(os.Stderr, "result is::  %+v\n", result)
	// fmt.Fprintf(os.Stderr, "result err is::  %+v\n", err)
	if err != nil {
		return nil, err
	}
	if result != nil {
		result = compactMap(result)
		// fmt.Fprintf(os.Stderr, "result after compact is::  %+v\n", result)
		return result, nil
	}
	return nil, nil
}

func toMap(str string) map[string]string {
	result := make(map[string]string)
	if str == "" {
		return result
	}
	vals := strings.Split(str, ",")
	for i := 0; i < len(vals); i++ {
		result[vals[i]] = vals[i]
	}
	return result
}

func ReleaseOthers(deployments string) error {
	lock := trylock()
	defer lock.Unlock()
	deps := toMap(deployments)
	if len(deps) == 0 {
		return fmt.Errorf("no deployments found, exit release for %+v", deployments)
	}

	a := &Persist{}
	a.reset()
	err := a.init(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(a.Inputs); i++ {
		for n := 0; n < len(a.Inputs[i].Used); n++ {
			used := &(a.Inputs[i].Used[n])
			if deps[used.Deployment] == "" {
				a.Inputs[i].Used = append(a.Inputs[i].Used[0:n], a.Inputs[i].Used[n+1:]...)
				a.Inputs[i].Available.push(used.Values, used.Group)
				// fmt.Println(fmt.Sprintf("fater release for ----   %+v", a.Inputs[i]))
			}
		}
	}
	a.save()
	return nil

}
func Release(deployment string) error {
	lock := trylock()
	defer lock.Unlock()

	a := &Persist{}
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
				a.Inputs[i].Available.push(used.Values, used.Group)
				// fmt.Println(fmt.Sprintf("fater release for ----   %+v", a.Inputs[i]))
			}
		}
	}
	// fmt.Println(fmt.Sprintf("fater release for ----   %+v", a))
	a.save()
	return nil
}
