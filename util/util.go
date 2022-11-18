package util

import (
	"encoding/json"
	"fmt"
)

func ByteToStruct(src []byte, des interface{}) error {
	err := json.Unmarshal(src, des)
	if err != nil {
		return fmt.Errorf("ByteToStruct json Unmarshal error : %v", err)
	}
	return nil
}

func InterfaceToStruct(src, des interface{}) error {
	if src == nil {
		return fmt.Errorf("InterfaceToStruct src is nil")
	}
	byteData, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("InterfaceToStruct json Marshal error: %v", err)
	}

	err = json.Unmarshal(byteData, des)
	if err != nil {
		return fmt.Errorf("InterfaceToStruct json Unmarshal error: %v", err)
	}

	return nil
}
