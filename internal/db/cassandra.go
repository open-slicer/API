package db

import (
	"github.com/gocql/gocql"
	"slicerapi/internal/util"
)

var Cassandra *gocql.Session

func ConnectCassandra() error {
	cluster := gocql.NewCluster(util.Config.DB.Cassandra.Hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: util.Config.DB.Cassandra.Username,
		Password: util.Config.DB.Cassandra.Password,
	}
	cluster.Keyspace = util.Config.DB.Cassandra.Keyspace
	cluster.Consistency = gocql.Quorum

	// Declaring this here as Cassandra wouldn't be assigned to.
	var err error
	Cassandra, err = cluster.CreateSession()
	return err
}
