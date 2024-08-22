package lib

func ContinueOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
