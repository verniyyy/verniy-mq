package util

import "github.com/oklog/ulid/v2"

// GenULID ...
func GenULID() string {
	return ulid.Make().String()
}
