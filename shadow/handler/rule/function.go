package rule

var _p *Rule

func Start() error {
	_p = newFule()
	err := _p.start()
	if err != nil {
		panic(err)
	}
	return nil
}
