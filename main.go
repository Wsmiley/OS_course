package main

// rsrc -manifest app.manifest -o rsrc.syso
//go build

import (
	"fmt"

	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MyMainWindow struct {
	*walk.MainWindow
	producer, consumer, buff *walk.TextEdit
	buff1, buff2, buff3      *walk.PushButton
	label1                   *walk.Label
}

// var wg sync.WaitGroup
var cond sync.Cond  // 创建全局条件变量
var buffarea [3]int //int array with length 3
var countP = 2

// walk.Resources.SetRootDirPath("./img")

// v, _ := mem.VirtualMemory()
// fmt.Printf("Total: %v, Free%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

func main() {
	rand.Seed(time.Now().UnixNano()) // 设置随机数种子
	quit := make(chan bool)          // 创建用于结束通的 channel
	product := make(chan int, 3)     // 产品区（公共区）使用channel模拟
	cond.L = new(sync.Mutex)         // 创建互斥锁和条件变量
	var n = 0
	var j = 0

	mw := new(MyMainWindow)

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "生产者消费者问题",
		MinSize:  Size{270, 20},
		Size:     Size{1000, 500},
		Layout:   HBox{},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					PushButton{
						Text:    "加生产者",
						MinSize: Size{100, 50},
						OnClicked: func() {
							n++
							mw.producer.AppendText("\r\n" + strconv.Itoa(n) + "th 生产者")
							go producer(product, n, mw) //开启一个生产者协程
						},
					},
					PushButton{
						Text:    "加消费者",
						MinSize: Size{100, 50},
						OnClicked: func() {
							j++
							mw.consumer.AppendText("\r\n" + strconv.Itoa(j) + "th 消费者")
							go consumer(product, j, mw) //开启一个消费者协程
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{Text: "剩余缓冲区:"},
							Label{
								AssignTo: &mw.label1,
								Text:     "3",
							},
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "*************生产者线程**************"},
							TextEdit{AssignTo: &mw.producer, VScroll: true},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "*************缓冲区****************"},
							TextEdit{AssignTo: &mw.buff, VScroll: true},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										AssignTo: &mw.buff1,
										MinSize:  Size{10, 10},
									},
									PushButton{
										AssignTo: &mw.buff2,
										MinSize:  Size{10, 10},
									},
									PushButton{
										AssignTo: &mw.buff3,
										MinSize:  Size{10, 10},
									},
								},
							},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "*************消费者线程**************"},
							TextEdit{AssignTo: &mw.consumer, VScroll: true},
						},
					},
				},
			},
		},
	}.Create()); err != nil {

	}
	mw.Run()
	<-quit // 主协程阻塞 不结束
}

//  生产者
func producer(out chan<- int, idx int, mw *MyMainWindow) {
	for {
		cond.L.Lock()       // 条件变量对应互斥锁加锁
		for len(out) == 3 { // 产品区满 等待消费者消
			mw.buff.AppendText("\r\n" + "缓冲区已满生成者" + strconv.Itoa(idx) + "等待")
			cond.Wait() // 挂起当前协程， 等待条件变量满足，被消费者唤醒
		}

		num := rand.Intn(1000) // 产生一个随机数
		out <- num             // 写入到 chanel 中 （生产）
		buffarea[countP] = num
		buffnum := len(out)
		switch buffnum {
		case 1:
			{
				mw.buff3.SetText(strconv.Itoa(buffarea[2]))
				mw.label1.SetText(strconv.Itoa(3 - len(out)))
			}
		case 2:
			{
				mw.buff3.SetText(strconv.Itoa(buffarea[2]))
				mw.buff2.SetText(strconv.Itoa(buffarea[1]))
				mw.label1.SetText(strconv.Itoa(3 - len(out)))

			}
		case 3:
			{
				mw.buff3.SetText(strconv.Itoa(buffarea[2]))
				mw.buff2.SetText(strconv.Itoa(buffarea[1]))
				mw.buff1.SetText(strconv.Itoa(buffarea[0]))
				mw.label1.SetText(strconv.Itoa(3 - len(out)))
			}
		}
		mw.buff.AppendText("\r\n" + strconv.Itoa(idx) + "th 生产者,产生数据 " + strconv.Itoa(num) +
			",公共区有" + strconv.Itoa(len(out)) + "个数据\n")
		fmt.Printf("%dth 生产者，产生数据 %3d, 公共区剩余%d个数据\n", idx, num, len(out))
		countP--
		if countP < 0 {
			countP = 2
		}
		cond.L.Unlock()         // 生产结束，解锁互斥锁
		cond.Signal()           // 唤醒 阻塞的 消费者
		time.Sleep(time.Second) // 生产完休息一会，给其协程执行机会
	}
}

// 消费者
func consumer(in <-chan int, idx int, mw *MyMainWindow) {
	for {
		cond.L.Lock()
		for len(in) == 0 { // 产品区为空 等待生产者生产
			mw.buff.AppendText("\r\n" + "缓冲区为空消费者" + strconv.Itoa(idx) + "等待")
			cond.Wait() // 挂起当前协程， 等待条件变量满足，被生产者唤醒
		}

		num := <-in // 将 channel 中的数据读走 （消费）
		buffnum := len(in)
		switch buffnum {
		case 0:
			{
				mw.buff1.SetText("")
				mw.buff2.SetText("")
				mw.buff3.SetText("")
				mw.label1.SetText(strconv.Itoa(3 - len(in)))

			}
		case 1:
			{
				mw.buff3.SetText(strconv.Itoa(buffarea[2]))
				mw.buff2.SetText("")
				mw.buff1.SetText("")
				mw.label1.SetText(strconv.Itoa(3 - len(in)))

			}
		case 2:
			{
				buffarea[2] = buffarea[1]
				buffarea[1] = buffarea[0]
				mw.buff3.SetText(strconv.Itoa(buffarea[2]))
				mw.buff2.SetText(strconv.Itoa(buffarea[1]))
				mw.buff1.SetText("")
				mw.label1.SetText(strconv.Itoa(3 - len(in)))
			}
		}
		mw.buff.AppendText("\r\n" + strconv.Itoa(idx) + "th 消费者, 消费数据 " + strconv.Itoa(num) +
			",公共区剩余" + strconv.Itoa(len(in)) + "个数据\n")
		fmt.Printf("---- %dth 消费者, 消费数据 %3d,公共区剩余%d个数据\n", idx, num, len(in))
		countP++
		if countP > 2 {
			countP = 0
		}
		cond.L.Unlock()                    // 消费结束，解锁互斥锁
		cond.Signal()                      // 唤醒 阻塞的 生产者
		time.Sleep(time.Millisecond * 500) //消费完 休息一会，给其协程执行机会

	}
}
