package db

import (
	"slicerapi/internal/util"

	"github.com/gocql/gocql"
)

// Cassandra is the main gocql session used by the app.
var Cassandra *gocql.Session

// Connect creates a gocql session and assigns Cassandra to it.
func Connect() error {
	cluster := gocql.NewCluster(util.Config.DB.Cassandra.Hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: util.Config.DB.Cassandra.Username,
		Password: util.Config.DB.Cassandra.Password,
	}
	cluster.Keyspace = util.Config.DB.Cassandra.Keyspace
	cluster.Consistency = gocql.One

	// Declaring this here as Cassandra wouldn't be assigned to.
	var err error
	Cassandra, err = cluster.CreateSession()
	return err
}
