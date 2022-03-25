package main

import (
	"fmt"
	hook "github.com/robotn/gohook"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	w = 0
	h = 0
)

const (
	food      = 3
	head      = 2
	body      = 1
	baseSpeed = 100
)

var bodyArray []point //蛇点位记录数组
var integral = 0      //积分
var vector = int64(6) //方向 根据小数字键盘 2下 4左 6右 8上 方向来
var speed = 400       //附加间隔时间
var speedUp = false   //是否加速

type point struct {
	x    int
	y    int
	Type int
}

var foodP point

func main() {
	println("请回车开始")
	var a int
	_, _ = fmt.Scanln(&a)
	rand.Seed(time.Now().Unix())
	iniBody()
	randFood()
	go listeningInput()
	for {
		if !next() {
			println("game over")
			return
		}
		drawMap()
		if speedUp {
			time.Sleep(time.Millisecond * time.Duration(baseSpeed))
		} else {
			time.Sleep(time.Millisecond * time.Duration(baseSpeed+speed))
		}
	}
}

func listeningInput() {
	EvChan := hook.Start()
	defer hook.StopEvent()
	for ev := range EvChan {
		if ev.Kind == hook.KeyHold && ev.Keycode == 57 {
			speedUp = true
		}
		if ev.Kind == hook.KeyUp && ev.Keycode == 57 {
			speedUp = false
		}
		if ev.Kind == hook.KeyHold {
			//fmt.Println(ev.Keycode, ev.Rawcode, ev)
			switch ev.Keycode {
			case 17, 61000, 57416:
				if atomic.LoadInt64(&vector) != 2 {
					atomic.StoreInt64(&vector, 8)
				}
			case 30, 61003, 57419:
				if atomic.LoadInt64(&vector) != 6 {
					atomic.StoreInt64(&vector, 4)
				}
			case 31, 61008, 57424:
				if atomic.LoadInt64(&vector) != 8 {
					atomic.StoreInt64(&vector, 2)
				}
			case 32, 61005, 57421:
				if atomic.LoadInt64(&vector) != 4 {
					atomic.StoreInt64(&vector, 6)
				}
			}
		}
	}
}
func randFood() {
	x := rand.Int63n(int64(w))
	y := rand.Int63n(int64(h))
	flg := false
	for _, p := range bodyArray {
		if int64(p.x) == x && int64(p.y) == y {
			flg = true
		}
	}
	if flg {
		randFood()
		speed -= 10
	} else {
		foodP.x = int(x)
		foodP.y = int(y)
		foodP.Type = food
	}
}
func next() bool {
	//蛇的移动是去掉最后一个点位 把头部点位移到新的点位
	newP := point{
		x:    bodyArray[0].x,
		y:    bodyArray[0].y,
		Type: bodyArray[0].Type,
	}
	//根据方向移动
	switch atomic.LoadInt64(&vector) {
	case 8:
		newP.y -= 1
	case 2:
		newP.y += 1
	case 4:
		newP.x -= 1
	case 6:
		newP.x += 1
	}
	//添加到头部
	bodyArray = append([]point{newP}, bodyArray...)
	//如果没有吃到食物就去掉尾巴
	if !(newP.x == foodP.x && newP.y == foodP.y) {
		bodyArray = bodyArray[:len(bodyArray)-1]
	} else {
		//吃到食物就添加积分 随机新的食物
		randFood()
		integral++
	}
	//改变原来头部类型为身体
	bodyArray[1].Type = body
	//超过边界 结束
	if newP.y > h || newP.y < 0 {
		return false
	}
	if newP.x > w || newP.x < 0 {
		return false
	}
	//撞到自己
	for _, p := range bodyArray[1:] {
		if bodyArray[0].x == p.x && bodyArray[0].y == p.y {
			return false
		}
	}
	return true
}
func iniBody() {
	var err error
	h, w, err = getWSZ()
	if err != nil {
		println(err.Error())
		return
	}
	h -= 1 //减去一行用来输出积分信息
	bodyArray = []point{
		{
			x:    2,
			y:    0,
			Type: head,
		},
		{
			x:    1,
			y:    0,
			Type: body,
		},
		{
			x:    0,
			y:    0,
			Type: body,
		},
	}
}
func drawBar() {
	fmt.Println(fmt.Sprintf("当前积分：%d", integral), speedUp)
}
func drawMap() {
	clear()
	drawBar()
	m := make(map[string]int)
	m[fmt.Sprintf("%d-%d", foodP.x, foodP.y)] = food
	for _, p := range bodyArray {
		m[fmt.Sprintf("%d-%d", p.x, p.y)] = p.Type
	}
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			k := fmt.Sprintf("%d-%d", j, i)
			out := " "
			if v, ok := m[k]; ok {
				switch v {
				case body:
					out = "*"
				case head:
					out = "@"
				case food:
					out = "A"
				}
			}
			print(out)
		}
	}
}

//获取终端宽高
func getWSZ() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	a := strings.ReplaceAll(strings.ReplaceAll(string(out), "\r", ""), "\n", "")
	split := strings.Split(a, " ")
	h, _ := strconv.Atoi(split[0])
	w, _ := strconv.Atoi(split[1])
	return h, w, err
}

//清空终端
func clear() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}
