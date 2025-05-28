package externals

type AllAppExternals struct {
	JsonDB *JsonDBExternal
}

func RegisterExternals() (*AllAppExternals, error) {
	// register db externals
	dbInstance := &JsonDBExternal{}

	dbInstance, dbInstanceError := dbInstance.Connect()

	if dbInstanceError == nil {
		dbInstanceErrorHealthCheck := dbInstance.HealthCheck()

		if dbInstanceErrorHealthCheck != nil {
			return nil, dbInstanceErrorHealthCheck
		}
	} else {
		return nil, dbInstanceError
	}

	// can do for other externals like Redis, MongoDB, etc.

	return &AllAppExternals{
		JsonDB: dbInstance,
	}, nil
}
