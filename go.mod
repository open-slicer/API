module slicerapi

go 1.14

require (
	github.com/SlicerChat/PGP v0.0.0-20200519183404-576a5b8dbcd4
	github.com/appleboy/gin-jwt/v2 v2.6.3
	github.com/gin-gonic/gin v1.6.3
	github.com/gocql/gocql v0.0.0-20200515162754-0714040f3e35
	github.com/gorilla/websocket v1.4.2
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace golang.org/x/crypto => github.com/ProtonMail/crypto v0.0.0-20200416114516-1fa7f403fb9c
