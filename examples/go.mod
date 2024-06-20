module github.com/bartventer/gorm-multitenancy/examples/v6

go 1.22

toolchain go1.22.4

replace github.com/bartventer/gorm-multitenancy/v6 => ../

replace github.com/bartventer/gorm-multitenancy/drivers/postgres/v6 => ../drivers/postgres

replace github.com/bartventer/gorm-multitenancy/middleware/echo/v6 => ../middleware/echo

replace github.com/bartventer/gorm-multitenancy/middleware/nethttp/v6 => ../middleware/nethttp

require (
	github.com/bartventer/gorm-multitenancy/drivers/postgres/v6 v6.0.0-00010101000000-000000000000
	github.com/bartventer/gorm-multitenancy/middleware/echo/v6 v6.0.0-00010101000000-000000000000
	github.com/bartventer/gorm-multitenancy/middleware/nethttp/v6 v6.0.0-00010101000000-000000000000
	github.com/labstack/echo/v4 v4.12.0
	github.com/urfave/negroni v1.0.0
	gorm.io/gorm v1.25.10
)

require (
	github.com/bartventer/gorm-multitenancy/v6 v6.1.3 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gorm.io/driver/postgres v1.5.9 // indirect
)
