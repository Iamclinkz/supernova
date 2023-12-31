package util

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// 魔改的大致from：https://github.com/rfyiamcool/go-timewheel
const (
	typeTimer taskType = iota
	typeTicker

	modeIsCircle  = true
	modeNotCircle = false

	modeIsAsync  = true
	modeNotAsync = false
)

type taskType int64
type taskID int64

type Task struct {
	delay    time.Duration
	id       taskID
	round    int
	callback func()

	async  bool
	stop   bool
	circle bool
	// circleNum int
}

// for sync.Pool
func (t *Task) Reset() {
	t.delay = 0
	t.id = 0
	t.round = 0
	t.callback = nil

	t.async = false
	t.stop = false
	t.circle = false
}

type optionCall func(*TimeWheel) error

func TickSafeMode() optionCall {
	return func(o *TimeWheel) error {
		o.tickQueue = make(chan time.Time, 100)
		return nil
	}
}

type TimeWheel struct {
	randomID int64

	tick      time.Duration
	ticker    *time.Ticker
	tickQueue chan time.Time

	bucketsNum    int
	buckets       []map[taskID]*Task // 时间轮中的所有格子，每个格子是一个 map，用于存储在该格子中需要执行的任务。
	bucketIndexes map[taskID]int     // 用于记录每个任务所在的格子的索引。

	currentIndex int

	onceStart sync.Once

	stopC chan struct{}

	exited bool

	sync.RWMutex
}

// NewTimeWheel create new time wheel
func NewTimeWheel(tick time.Duration, bucketsNum int, options ...optionCall) (*TimeWheel, error) {
	if tick.Milliseconds() < 1 {
		return nil, errors.New("invalid params, must tick >= 1 ms")
	}
	if bucketsNum <= 0 {
		return nil, errors.New("invalid params, must bucketsNum > 0")
	}

	tw := &TimeWheel{
		// tick
		tick:      tick,
		tickQueue: make(chan time.Time, 10),

		// store
		bucketsNum:    bucketsNum,
		bucketIndexes: make(map[taskID]int, 1024*100),
		buckets:       make([]map[taskID]*Task, bucketsNum),
		currentIndex:  0,

		// signal
		stopC: make(chan struct{}),
	}

	for i := 0; i < bucketsNum; i++ {
		tw.buckets[i] = make(map[taskID]*Task, 16)
	}

	for _, op := range options {
		op(tw)
	}

	return tw, nil
}

// Start start the time wheel
func (tw *TimeWheel) Start() {
	// onlye once start
	tw.onceStart.Do(
		func() {
			tw.ticker = time.NewTicker(tw.tick)
			go tw.schduler()
			go tw.tickGenerator()
		},
	)
}

func (tw *TimeWheel) tickGenerator() {
	if tw.tickQueue == nil {
		return
	}

	for !tw.exited {
		select {
		case <-tw.ticker.C:
			select {
			case tw.tickQueue <- time.Now():
			default:
				panic("raise long time blocking")
			}
		}
	}
}

func (tw *TimeWheel) schduler() {
	queue := tw.ticker.C
	if tw.tickQueue != nil {
		queue = tw.tickQueue
	}

	for {
		select {
		case <-queue:
			tw.handleTick()

		case <-tw.stopC:
			tw.exited = true
			tw.ticker.Stop()
			return
		}
	}
}

// Stop stop the time wheel
func (tw *TimeWheel) Stop() {
	tw.stopC <- struct{}{}
}

func (tw *TimeWheel) collectTask(task *Task) bool {
	index, ok := tw.bucketIndexes[task.id]
	if !ok {
		return false
	}
	delete(tw.bucketIndexes, task.id)
	delete(tw.buckets[index], task.id)
	return true
}

func (tw *TimeWheel) handleTick() {
	tw.Lock()
	defer tw.Unlock()

	bucket := tw.buckets[tw.currentIndex]
	for k, task := range bucket {
		if task.stop {
			tw.collectTask(task)
			continue
		}

		if bucket[k].round > 0 {
			bucket[k].round--
			continue
		}

		if task.async {
			go task.callback()
		} else {
			// optimize gopool
			task.callback()
		}

		// circle
		if task.circle == true {
			tw.collectTask(task)
			tw.putCircle(task, modeIsCircle)
			continue
		}

		// gc
		tw.collectTask(task)
	}

	if tw.currentIndex == tw.bucketsNum-1 {
		tw.currentIndex = 0
		return
	}

	tw.currentIndex++
}

// Add add an task
func (tw *TimeWheel) Add(delay time.Duration, callback func(), async bool) *Task {
	return tw.addAny(delay, callback, modeNotCircle, async)
}

// AddCron add interval task
func (tw *TimeWheel) AddCron(delay time.Duration, callback func(), async bool) *Task {
	return tw.addAny(delay, callback, modeIsCircle, async)
}

func (tw *TimeWheel) addAny(delay time.Duration, callback func(), circle, async bool) *Task {
	if delay <= 0 {
		delay = tw.tick
	}

	id := tw.genUniqueID()
	task := new(Task)

	task.delay = delay
	task.id = id
	task.callback = callback
	task.circle = circle
	task.async = async

	klog.Tracef("add task to timeWheel:%+v", task)

	tw.put(task)
	return task
}

func (tw *TimeWheel) put(task *Task) {
	tw.Lock()
	defer tw.Unlock()

	tw.store(task, false)
}

func (tw *TimeWheel) putCircle(task *Task, circleMode bool) {
	tw.store(task, circleMode)
}

func (tw *TimeWheel) store(task *Task, circleMode bool) {
	round := tw.calculateRound(task.delay)
	index := tw.calculateIndex(task.delay)

	if round > 0 && circleMode {
		task.round = round - 1
	} else {
		task.round = round
	}

	tw.bucketIndexes[task.id] = index
	tw.buckets[index][task.id] = task
}

func (tw *TimeWheel) calculateRound(delay time.Duration) (round int) {
	delaySeconds := delay.Seconds()
	tickSeconds := tw.tick.Seconds()
	round = int(delaySeconds / tickSeconds / float64(tw.bucketsNum))
	return
}

func (tw *TimeWheel) calculateIndex(delay time.Duration) (index int) {
	delaySeconds := delay.Seconds()
	tickSeconds := tw.tick.Seconds()
	index = (int(float64(tw.currentIndex) + delaySeconds/tickSeconds)) % tw.bucketsNum
	return
}

func (tw *TimeWheel) Remove(task *Task) bool {
	// tw.removeC <- task
	return tw.remove(task)
}

func (tw *TimeWheel) remove(task *Task) bool {
	tw.Lock()
	defer tw.Unlock()

	return tw.collectTask(task)
}

func (tw *TimeWheel) NewTimer(delay time.Duration) *Timer {
	queue := make(chan bool, 1) // buf = 1, refer to src/time/sleep.go
	task := tw.addAny(delay,
		func() {
			notifyChannel(queue)
		},
		modeNotCircle,
		modeNotAsync,
	)

	// init timer
	ctx, cancel := context.WithCancel(context.Background())
	timer := &Timer{
		tw:     tw,
		C:      queue, // faster
		task:   task,
		Ctx:    ctx,
		cancel: cancel,
	}

	return timer
}

func (tw *TimeWheel) AfterFunc(delay time.Duration, callback func()) *Timer {
	queue := make(chan bool, 1)
	task := tw.addAny(delay,
		func() {
			callback()
			notifyChannel(queue)
		},
		modeNotCircle, modeIsAsync,
	)

	// init timer
	ctx, cancel := context.WithCancel(context.Background())
	timer := &Timer{
		tw:     tw,
		C:      queue, // faster
		task:   task,
		Ctx:    ctx,
		cancel: cancel,
		fn:     callback,
	}

	return timer
}

func (tw *TimeWheel) NewTicker(delay time.Duration) *Ticker {
	queue := make(chan bool, 1)
	task := tw.addAny(delay,
		func() {
			notifyChannel(queue)
		},
		modeIsCircle,
		modeNotAsync,
	)

	// init ticker
	ctx, cancel := context.WithCancel(context.Background())
	ticker := &Ticker{
		task:   task,
		tw:     tw,
		C:      queue,
		Ctx:    ctx,
		cancel: cancel,
	}

	return ticker
}

func (tw *TimeWheel) After(delay time.Duration) <-chan time.Time {
	queue := make(chan time.Time, 1)
	tw.addAny(delay,
		func() {
			queue <- time.Now()
		},
		modeNotCircle, modeNotAsync,
	)
	return queue
}

func (tw *TimeWheel) Sleep(delay time.Duration) {
	queue := make(chan bool, 1)
	tw.addAny(delay,
		func() {
			queue <- true
		},
		modeNotCircle, modeNotAsync,
	)
	<-queue
}

// similar to golang std timer
type Timer struct {
	task   *Task
	tw     *TimeWheel
	fn     func() // external custom func
	stopFn func() // call function when timer stop

	C chan bool

	cancel context.CancelFunc
	Ctx    context.Context
}

func (t *Timer) Reset(delay time.Duration) {
	// first stop old task
	t.task.stop = true

	// make new task
	var task *Task
	if t.fn != nil { // use AfterFunc
		task = t.tw.addAny(delay,
			func() {
				t.fn()
				notifyChannel(t.C)
			},
			modeNotCircle, modeIsAsync, // must async mode
		)
	} else {
		task = t.tw.addAny(delay,
			func() {
				notifyChannel(t.C)
			},
			modeNotCircle, modeNotAsync)
	}

	t.task = task
}

func (t *Timer) Stop() {
	if t.stopFn != nil {
		t.stopFn()
	}

	t.task.stop = true
	t.cancel()
	t.tw.Remove(t.task)
}

func (t *Timer) AddStopFunc(callback func()) {
	t.stopFn = callback
}

type Ticker struct {
	tw     *TimeWheel
	task   *Task
	cancel context.CancelFunc

	C   chan bool
	Ctx context.Context
}

func (t *Ticker) Stop() {
	t.task.stop = true
	t.cancel()
	t.tw.Remove(t.task)
}

func notifyChannel(q chan bool) {
	select {
	case q <- true:
	default:
	}
}

func (tw *TimeWheel) genUniqueID() taskID {
	id := atomic.AddInt64(&tw.randomID, 1)
	return taskID(id)
}
