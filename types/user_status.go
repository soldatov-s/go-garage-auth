package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

var (
	ErrBadStatus = errors.New("bad status")
)

type Status int

const (
	New Status = iota // default value
	Active
	Restricted
)

func StringToStatus() map[string]Status {
	return map[string]Status{
		"NEW":        New,
		"ACTIVE":     Active,
		"RESTRICTED": Restricted,
	}
}

func (status Status) String() string {
	for key, item := range StringToStatus() {
		if item == status {
			return key
		}
	}

	return ""
}

// UnmarshalJSON method is called by json.Unmarshal,
// whenever it is of type Status
func (status *Status) UnmarshalJSON(data []byte) error {
	var statusName string

	if data == nil {
		*status = New
		return nil
	}

	if err := json.Unmarshal(data, &statusName); err != nil {
		return err
	}

	// Check received Role
	if statusName == "" {
		*status = New
	} else {
		s, ok := StringToStatus()[statusName]
		if !ok {
			return ErrBadStatus
		}
		*status = s
	}

	return nil
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type Status
func (status *Status) MarshalJSON() ([]byte, error) {
	statusName := status.String()

	if statusName == "" {
		return nil, ErrBadStatus
	}

	return json.Marshal(statusName)
}

// Value implements the driver Valuer interface.
func (status Status) Value() (driver.Value, error) {
	statusName := status.String()

	if statusName == "" {
		return nil, ErrBadStatus
	}

	return statusName, nil
}

// Make the Status struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (status *Status) Scan(value interface{}) error {
	if value == nil {
		*status = New
		return nil
	}

	b, ok := value.(string)

	if !ok {
		return errors.New("type assertion to string failed")
	}

	*status = StringToStatus()[b]

	return nil
}
