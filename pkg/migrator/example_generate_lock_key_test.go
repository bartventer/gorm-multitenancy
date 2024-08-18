package migrator_test

import (
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
)

func ExampleGenerateLockKey() {
	key := migrator.GenerateLockKey("test")
	fmt.Printf("%d", key)
	// Output: -439409999022904539
}
