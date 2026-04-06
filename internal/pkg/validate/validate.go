package validate

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func Required(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

func UUID(value string, field string) error {
	if err := Required(value, field); err != nil {
		return err
	}
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%s must be uuid: %w", field, err)
	}
	return nil
}

func NonNegativeInt(value int, field string) error {
	if value < 0 {
		return fmt.Errorf("%s must be >= 0", field)
	}
	return nil
}

