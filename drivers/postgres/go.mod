module github.com/bartventer/gorm-multitenancy/drivers/postgres/v6

go 1.22

toolchain go1.22.4

replace github.com/bartventer/gorm-multitenancy/v6 => ../..

require (
	github.com/bartventer/gorm-multitenancy/v6 v6.1.3
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.10
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)
