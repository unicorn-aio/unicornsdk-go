package unicornsdk

type ParamsInvalid struct {
	Errmsg string
}

func (self ParamsInvalid) Error() string {
	return self.Errmsg
}

type NotAuthenticated struct {
	Errmsg string
}

func (self NotAuthenticated) Error() string {
	return self.Errmsg
}
