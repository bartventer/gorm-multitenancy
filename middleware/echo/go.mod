module github.com/bartventer/gorm-multitenancy/middleware/echo/v8

go 1.23

replace github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8 => ../nethttp

require (
	github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8 v8.7.1
	github.com/labstack/echo/v4 v4.12.0
)

require (
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)
