package db

// Connect connects to both Redis and Cassandra.
func Connect() error {
	if err := ConnectRedis(); err != nil {
		return err
	}
	if err := ConnectCassandra(); err != nil {
		return err
	}

	return nil
}
