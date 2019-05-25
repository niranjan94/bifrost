package utils

// Must panics if the error is not nil
func Must(err error)  {
	if err != nil {
		panic(err)
	}
}
