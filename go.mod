module slicerapi

go 1.14

require (
	github.com/appleboy/gin-jwt/v2 v2.6.3
	github.com/gin-gonic/gin v1.6.3
	github.com/gocql/gocql v0.0.0-20200608162118-cb62e193e52b
	github.com/gorilla/websocket v1.4.2
	github.com/sirupsen/logrus v1.6.0
	go.mongodb.org/mongo-driver v1.3.4
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace golang.org/x/crypto => github.com/ProtonMail/crypto v0.0.0-20200416114516-1fa7f403fb9c
