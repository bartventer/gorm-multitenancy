module github.com/bartventer/gorm-multitenancy/examples/v8

go 1.24

replace github.com/bartventer/gorm-multitenancy/v8 => ../

replace github.com/bartventer/gorm-multitenancy/postgres/v8 => ../postgres

replace github.com/bartventer/gorm-multitenancy/mysql/v8 => ../mysql

replace github.com/bartventer/gorm-multitenancy/middleware/echo/v8 => ../middleware/echo

replace github.com/bartventer/gorm-multitenancy/middleware/gin/v8 => ../middleware/gin

replace github.com/bartventer/gorm-multitenancy/middleware/iris/v8 => ../middleware/iris

replace github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8 => ../middleware/nethttp

require (
	github.com/bartventer/gorm-multitenancy/middleware/echo/v8 v8.8.1
	github.com/bartventer/gorm-multitenancy/middleware/gin/v8 v8.5.1
	github.com/bartventer/gorm-multitenancy/middleware/iris/v8 v8.1.1
	github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8 v8.8.1
	github.com/bartventer/gorm-multitenancy/mysql/v8 v8.8.1
	github.com/bartventer/gorm-multitenancy/postgres/v8 v8.8.1
	github.com/bartventer/gorm-multitenancy/v8 v8.8.1
	github.com/docker/go-connections v0.5.0
	github.com/gin-gonic/gin v1.10.1
	github.com/kataras/iris/v12 v12.2.11
	github.com/labstack/echo/v4 v4.13.4
	github.com/stretchr/testify v1.10.0
	github.com/testcontainers/testcontainers-go v0.37.0
	github.com/urfave/negroni v1.0.0
	gorm.io/gorm v1.30.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/CloudyKit/fastprinter v0.0.0-20200109182630-33d98a066a53 // indirect
	github.com/CloudyKit/jet/v6 v6.3.1 // indirect
	github.com/Joker/jade v1.1.3 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Shopify/goreferrer v0.0.0-20250617153402-88c1d9a79b05 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v28.3.2+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/flosch/pongo2/v4 v4.0.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/gomarkdown/markdown v0.0.0-20250311123330-531bef5e742b // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/iris-contrib/schema v0.0.6 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kataras/blocks v0.0.11 // indirect
	github.com/kataras/golog v0.1.13 // indirect
	github.com/kataras/pio v0.0.14 // indirect
	github.com/kataras/sitemap v0.0.6 // indirect
	github.com/kataras/tunnel v0.0.4 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mailgun/raymond/v2 v2.0.48 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/schollz/closestmatch v2.1.0+incompatible // indirect
	github.com/shirou/gopsutil/v4 v4.25.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tdewolff/minify/v2 v2.23.8 // indirect
	github.com/tdewolff/parse/v2 v2.8.1 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/yosssi/ace v0.0.5 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

require (
	github.com/bytedance/sonic v1.13.3 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/fatih/color v1.18.0
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/arch v0.19.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
)
