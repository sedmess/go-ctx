package ctx

func ServiceArray(srvs ...Service) []Service {
	return srvs
}

func runWithRecover(block func(), onErr func(panicReason any)) {
	defer func() {
		if reason := recover(); reason != nil {
			onErr(reason)
		}
	}()
	block()
}
