module github.com/bartventer/gorm-multitenancy/middleware/echo/v8

go 1.24

replace github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8 => ../nethttp

require (
	github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8 v8.7.2
	github.com/labstack/echo/v4 v4.13.3
)

require (
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
