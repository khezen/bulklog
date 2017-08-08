package uuid

import (
	"fmt"
	"github.com/google/uuid"
)

// New generates a random UUID
func New() string {
	uuid := uuid.New()
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
