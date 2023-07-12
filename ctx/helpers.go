package ctx

func ServiceArray(srvs ...any) []any {
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
