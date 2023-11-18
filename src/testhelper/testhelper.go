package testhelper

import "fmt"

func EqualError(err1, err2 error) bool {
	return fmt.Sprint(err1) == fmt.Sprint(err2)
}
