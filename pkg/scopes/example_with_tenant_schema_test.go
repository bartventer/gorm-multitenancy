package scopes_test

import (
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/scopes"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
)

type Book struct {
	ID    uint
	Title string
}

func (Book) TableName() string   { return "books" }
func (Book) IsSharedModel() bool { return true }

func ExampleWithTenantSchema() {
	db, err := gorm.Open(tests.DummyDialector{})
	if err != nil {
		panic(err)
	}

	queryFn := func(tx *gorm.DB) *gorm.DB {
		return tx.Scopes(scopes.WithTenantSchema("tenant1")).Find(&Book{})
	}

	sqlstr := db.ToSQL(queryFn)
	fmt.Println(sqlstr)

	// Output:
	// SELECT * FROM `tenant1`.`books`
}
