package main

import (
	"fmt"
	"net"
	"reflect"
	"strings"
)

type QQ struct {
	b string
}

type PP struct {
	a  int
	b  string
	cs []QQ
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	fmt.Println(fmt.Sprintf("ip in inc is: %v", ip))
}

func combine(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	for key, value := range b {
		a[key] = value
	}
	return a

}
func toArray(prefix string, val map[string]interface{}) map[string]interface{} {
	fmt.Println(fmt.Sprintf("kkkkkkkkkkkkk %v ", prefix))
	result := map[string]interface{}{}
	for key, value := range val {
		if reflect.TypeOf(value).Kind() == reflect.Map {
			result = combine(result, toArray(prefix+key+".", (value.(map[string]interface{}))))
		} else {
			(result)[prefix+key] = value
		}
	}
	return result

}

func getValue(val map[string]interface{}, keys []string) map[string]interface{} {
	data := val
	data = val[keys[0]].(map[string]interface{})
	return data
}

func compactValue(val map[string]interface{}) map[string]interface{} {
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

func main1() {
	// v := PP{}
	// v.a = 10
	// v.cs = make([]QQ, 4)
	// for n := 0; n < 4; n++ {
	// 	v.cs[n].b = fmt.Sprintf("init %d", n)
	// }
	// for i, c := range v.cs {
	// 	c.b = fmt.Sprintf("haha %d", i)
	// }
	//
	// for i0, c0 := range v.cs {
	// 	c0.b = fmt.Sprintf("kk %d", i0)
	// }
	//
	// fmt.Printf("k value is: %v", v)
	// ff := " 1234kkk "
	// mm := fmt.Sprintf(" 1234%s ", "kkk")
	//
	// fmt.Printf("string compare : %v", ff == mm)
	// var ll []string
	// fmt.Printf("array length is : %v", len(ll))
	// qq := []string{"1", "2", "3"}
	// fmt.Printf("array append : %v", append(ll, qq...))
	// ip := net.ParseIP("192.168.1.0")
	// ip1 := net.ParseIP("192.168.1.1")
	// inc(ip)
	// fmt.Println(fmt.Sprintf("1 : %v", ip))
	// fmt.Println(fmt.Sprintf("2 : %v", ip1))

	// pp := []string{"a", "b", "c"}
	// pp = append(pp[0:2], pp[3:]...)
	// fmt.Println(fmt.Sprintf("2 : %v", pp))

	// oo := map[string]interface{}{"a.b.c": "c0", "a.b.d": "dd", "a.oo": "oo0", "q": "q0"}
	// pp := map[string]interface{}{"q": "q0"}
	// ff := map[string]interface{}{"c": "c0"}
	// pp["b"] = ff
	//
	// fmt.Println(fmt.Sprintf(".........kkk   %v ", pp))
	// fmt.Println(fmt.Sprintf(".........kkk   %v ", compactValue(oo)))
	//
	// var kk = "metadata.ddfd"
	// fmt.Println(fmt.Sprintf("value for print is::: %v ", strings.Split(kk, "metadata.")[1]))

	// match, _ := regexp.MatchString("^ *[0-9]*p([a-z]+)ch", "peach")
	// fmt.Println(match)
	// r, _ := regexp.Compile("^ *([0-9]*) *: *auto *$")
	// r, _ := regexp.Compile("(.*)\\[([^\\[\\]]*)\\]\\.*(.*)")
	// fmt.Println(r.MatchString("peach"))
	// fmt.Println(r.FindStringSubmatch(" pp[kk]")[3])
	// fmt.Println(r.FindStringSubmatch(" pp[kk]"))
	pp := make(map[string]interface{})
	// fmt.Println(len(r.FindStringSubmatch(" pp[kk]")))
	fmt.Println(fmt.Sprintf("hahaha %+v ", pp["kk"].(string)+"ddd"))
	// pp["kk"] = "ccc"
	// fmt.Println(fmt.Sprintf("hahaha %+v ", pp["kk"] == nil))
	// fmt.Println(r.FindStringIndex("peach punch"))
	// fmt.Println(r.FindStringSubmatch("peach punch"))
	// fmt.Println(r.FindStringSubmatchIndex("peach punch"))
	// fmt.Println(r.FindAllString("peach punch pinch", -1))
	// fmt.Println(r.FindAllStringSubmatchIndex(
	// 	"peach punch pinch", -1))
	// fmt.Println(r.FindAllString("peach punch pinch", 2))
	// fmt.Println(r.Match([]byte("peach")))
	// r = regexp.MustCompile("p([a-z]+)ch")
	// fmt.Println(r)
	// fmt.Println(r.ReplaceAllString("a peach", "<fruit>"))
	// in := []byte("a peach")
	// out := r.ReplaceAllFunc(in, bytes.ToUpper)
	// fmt.Println(string(out))

}
