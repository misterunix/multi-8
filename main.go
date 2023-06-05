package main

import (
	"fmt"
	v "multi-8/vm"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const screenWidth = 640
const screenHeight = 320

var vm v.VM

func main() {
	fmt.Println("Starting multi-8.")

	vm = v.New()

	vm.LoadProgram("roms/IBM.ch8")

	rl.InitWindow(screenWidth, screenHeight, "multi-8")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {
		update()
		draw()
	}

}

func update() {
	vm.ExecuteOpCode()
}

func draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	for y := 0; y < v.SCREENHEIGHT; y++ {
		for x := 0; x < v.SCREENWIDTH; x++ {
			ti := v.XYToIndex(uint8(x), uint8(y))
			if ti == 1 {
				rl.DrawRectangle(int32(x), int32(y), 1, 1, rl.White)
			}
		}
	}
}
