package gotrade

type Stragegy interface {
	Init(Account, Subscriber)
	Run()
}
