package models

import (
	"fmt"
	"time"
)

type RFC3339Time time.Time

func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	formattedTime := fmt.Sprintf(`"%s"`, time.Time(t).Format(time.RFC3339))

	return []byte(formattedTime), nil
}
