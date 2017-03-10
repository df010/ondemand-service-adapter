package persistent

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
)

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

func combine(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	for key, value := range b {
		a[key] = value
	}
	return a
}

func flatMap(prefix string, val map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range val {
		fmt.Fprintf(os.Stderr, "...................... %+v...... %+v", key, value)
		if value != nil {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				result = combine(result, flatMap(prefix+key+".", (value.(map[string]interface{}))))
			} else {
				(result)[prefix+key] = value
			}
		}
	}
	return result
}
