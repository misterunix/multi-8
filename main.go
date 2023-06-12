package main

import (
	"flag"
	"fmt"
	v "multi-8/vm"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const screenWidth = 640
const screenHeight = 320

var vm v.VM
var scale int32 = 10

var debug bool = false

func main() {
	fmt.Println("Starting multi-8.")

	flag.BoolVar(&debug, "debug", false, "Enable debug mode.")

	vm = v.New(debug)

	vm.LoadProgram("roms/caveexplorer.ch8")
	rl.SetTraceLog(rl.LogError)

	rl.InitWindow(screenWidth, screenHeight, "multi-8")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// run the vm here
	go func() {
		for {
			vm.ExecuteOpCode()
			time.Sleep(1 * time.Millisecond) // 1000hz
		}
	}()

	for !rl.WindowShouldClose() {
		update()
		draw()
	}

}

func update() {
	//vm.ExecuteOpCode()
}

func draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	for y := 0; y < v.SCREENHEIGHT; y++ {
		for x := 0; x < v.SCREENWIDTH; x++ {
			ti := v.XYToIndex(uint8(x), uint8(y))
			p := vm.Screen[ti]
			if p == 1 {
				rl.DrawRectangle(int32(x)*scale, int32(y)*scale, 1*scale, 1*scale, rl.White)
			} else {
				rl.DrawRectangle(int32(x)*scale, int32(y)*scale, 1*scale, 1*scale, rl.Black)
			}
		}
	}
	rl.EndDrawing()
}
