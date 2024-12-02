package parser

import (
	"fmt"
	"reflect"

	"github.com/oliveagle/jsonpath"
	"github.com/y7ut/potami/pkg/json"
)

// JsonPathOutputParser 解析JSON输出根据jsonPath规则
func JsonPathOutputParser(output []byte, jsonPaths map[string]string) (map[string]interface{}, error) {
	var outputSchema interface{}
	err := json.Unmarshal(output, &outputSchema)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for k, jp := range jsonPaths {
		res, err := jsonpath.JsonPathLookup(outputSchema, jp)
		if err != nil {
			return nil, err
		}
		if reflect.TypeOf(res).Kind() != reflect.String {
			res = fmt.Sprintf("%v is not string, only support string now", res)
		}
		result[k] = res
	}
	return result, nil
}
