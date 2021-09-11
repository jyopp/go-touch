package touch

var (
	runloop_tasks chan func()
)

func init() {
	runloop_tasks = make(chan func(), 100)
}

// AddToRunLoop adds a function for asynchronous execution in the main runloop
func AddToRunLoop(fn func()) {
	runloop_tasks <- fn
}
