// +build !linux

package byzanz

import (
	"fmt"
)

func SelectRectangle() (*Rectangle, error) {
	return nil, fmt.Errorf(`Rectangle is not supported other than Linux`)
}
