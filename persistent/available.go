package persistent

type Available struct {
	Arr []interface{}            `yaml:"arr"` // if groupby is nil
	Mp  map[string][]interface{} `yaml:"mp"`  //if groupby is not nil
}

func (ava *Available) push(arr []interface{}, groupby string) {
	if groupby == "" {
		ava.Arr = append(ava.Arr, arr...)
	} else {
		if ava.Mp == nil {
			ava.Mp = make(map[string][]interface{})
		}
		ava.Mp[groupby] = append(ava.Mp[groupby], arr...)
	}
}

func (ava *Available) set(arr []interface{}, groupby string) {
	if groupby == "" {
		ava.Arr = arr
	} else {
		if ava.Mp == nil {
			ava.Mp = make(map[string][]interface{})
		}
		ava.Mp[groupby] = arr
	}
}

func (ava *Available) avail(num int, groupby string) bool {
	if groupby == "" {
		return len(ava.Arr) >= num
	} else {
		if ava.Mp == nil {
			return num <= 0
		}
		return len(ava.Mp[groupby]) >= num
	}
}

func (ava *Available) pop(num int, groupby string) []interface{} {
	if groupby == "" {
		result := ava.Arr[0:num]
		ava.Arr = ava.Arr[num:]
		return result
	} else {
		if ava.Mp == nil {
			return nil
		}
		result := ava.Mp[groupby][0:num]
		ava.Mp[groupby] = ava.Mp[groupby][num:]
		return result
	}
}
