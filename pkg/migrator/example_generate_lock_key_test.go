package migrator_test

import (
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
)

func ExampleGenerateLockKey() {
	key := migrator.GenerateLockKey("test")
	fmt.Printf("%d", key)
	// Output: 18007334074686647077
}
