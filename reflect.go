/*
  反射相关的处理
  Autor: 不得闲
  QQ:75492895
*/

package DxCommonLib

import (
	"reflect"
	"strconv"
)

func IsSimpleCopyKind(kind reflect.Kind)bool  {
	switch kind {
	case reflect.String, reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	default:
		return false
	}
}

/*
field1->是否能够 匹配field2字段
 */
func CanConvertStructField(field1,field2 *reflect.StructField) bool {
	if field1.Type != field2.Type{
		f1Kind := field1.Type.Kind()
		if field1.Type.Kind() == reflect.Interface{
			return false
		}
		f2Kind := field2.Type.Kind()
		switch f1Kind {
		case reflect.Map:
			if f2Kind != reflect.Map && f2Kind != reflect.Struct{
				return false
			}
		case reflect.Struct:
			if f2Kind != reflect.Map && f2Kind != reflect.Struct{
				return false
			}
			if f2Kind == reflect.Map && field2.Type.Key().Kind() != reflect.String{
				return false
			}
		case reflect.Slice:
			if f2Kind == reflect.Slice {
				f1VKind := field1.Type.Elem().Kind()
				f2VKind := field2.Type.Elem().Kind()
				if f1VKind != f2VKind && f2VKind != reflect.Interface{
					return false
				}
			}
		default:
			return false
		}
	}
	if field1.Name == field2.Name{
		return true
	}
	//再判定tag
	tagMap1 := ParseStructTag(string(field1.Tag))
	if tagMap1 != nil{
		for _,v := range tagMap1{
			if v == field2.Name{
				return true
			}
		}
	}
	tagMap2 := ParseStructTag(string(field2.Tag))
	if tagMap2 != nil{
		for _,v := range tagMap2{
			if v == field1.Name{
				return true
			}
		}
	}
	if tagMap1 != nil && tagMap2 != nil{
		for _,v := range tagMap1{
			for _,v2 := range tagMap2{
				if v == v2{
					return true
				}
			}
		}
	}
	return false
}

//后续可以缓存起来

func ParseStructTag(tag string)map[string]string  {
	var result map[string]string
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]
		value, err := strconv.Unquote(qvalue)
		if err != nil {
			result = nil
			break
		}
		if result == nil{
			result = make(map[string]string)
		}
		result[name] = value
	}
	return result
}

/*
key是否匹配field
 */
func CanKeyMatchStructField(key string,field *reflect.StructField)bool  {
	if key == field.Name{
		return true
	}
	tagMap := ParseStructTag(string(field.Tag))
	if tagMap != nil{
		for _,v := range tagMap{
			if v == key{
				return true
			}
		}
	}
	return false
}