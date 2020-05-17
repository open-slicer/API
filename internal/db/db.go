package db

func Connect() error {
	if err := ConnectRedis(); err != nil {
		return err
	}
	if err := ConnectCassandra(); err != nil {
		return err
	}

	return nil
}
