package util

import "time"

//to avoid "decared but not used" or "imported but not used" when developing
func Use(...any) {}

func SetTimeout(f Callback, t time.Duration) {
	go func() {
		time.Sleep(t)
		SafeCallback(f)
	}()
}
