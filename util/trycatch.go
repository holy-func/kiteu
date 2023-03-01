package util


type TryCatchBody struct {
	Try     func()
	Catch   func(err any)
	Finally func()
}

func TryCatch(body *TryCatchBody) {
	try, catch, finally := body.Try, body.Catch, body.Finally
	NilDefend(try, catch)
	var err any = nil
	func() {
		defer func() {
			err = recover()
		}()
		try()
	}()
	if err != nil {
		catch(err)
	}
	SafeCallback(finally)
}