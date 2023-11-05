package datautils

import "encoding/json"

func ToJSONMap(data interface{}) (map[string]interface{}, error) {
	dataMap := map[string]interface{}{}

	// Marshal and unmarshal to/from JSON to get a map[string]interface{}. There are probably better
	// or more efficient ways to do this, but this is the easiest way for now.
	jsonBytes, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &dataMap)

	if err != nil {
		return nil, err
	}

	return dataMap, nil
}
