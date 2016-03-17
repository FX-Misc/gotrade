package gotrade

type Strategy interface {
	// 启动
	Run()
	// 暂停
	Pause()
	// 开始
	Start()
	// 状态
	Status() bool
}
