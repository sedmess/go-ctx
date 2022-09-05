package ctx

func ServiceArray(srvs ...Service) []Service {
	return srvs
}

func runWithRecover(block func(), onErr func(err error)) {
	defer func() {
		if err := recover(); err != nil {
			onErr(err.(error))
		}
	}()
	block()
}
