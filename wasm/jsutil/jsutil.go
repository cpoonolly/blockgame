package jsutil

import (
	"fmt"
	"syscall/js"
)

// ObjForEach iterates over a js object
// First param is the js object, second is a callback taking params val, key
// If the call back returns an error (non-nil value) the error is return by this func & execution is halted
func ObjForEach(obj js.Value, callback func(js.Value, string) error) error {
	if obj.Type() != js.TypeObject {
		return fmt.Errorf("attempt to iterate over non js object")
	}

	objKeys := js.Global().Get("Object").Call("keys", obj)
	objKeysLength := objKeys.Length()

	for i := 0; i < objKeysLength; i++ {
		key := objKeys.Index(i).String()
		val := obj.Get(key)

		if err := callback(val, key); err != nil {
			return err
		}
	}

	return nil
}
