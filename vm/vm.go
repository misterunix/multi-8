package vm

import (
	"fmt"
	"math/rand"
	"os"

	"time"
)

const (
	RAMSIZE      = 0x10000
	STACKSIZE    = 0x1E
	SCREENWIDTH  = 64
	SCREENHEIGHT = 32
	SCREENSIZE   = SCREENWIDTH * SCREENHEIGHT
)

type VM struct {
	I         uint16 // Index register
	Memory    [RAMSIZE]uint8
	Stack     [STACKSIZE]uint16
	PC        uint16
	SP        uint8
	Screen    [SCREENSIZE]uint8
	Timer     uint8
	Sound     uint8
	Keys      [16]uint8
	Registers [16]uint8
	OpCode    uint16     // 2 bytes opcode
	rnd       *rand.Rand // Random number generator
	x         uint8      // x register
	y         uint8      // y register
	n         uint8      // n nibble
	nn        uint8      // nn byte
	nnn       uint16     // nnn address
	debug     bool       // debug mode
}

func New(debug bool) VM {
	v := VM{}
	v.Init()
	v.debug = debug
	return v
}

func (v *VM) Reset() {
	fmt.Println("Resetting VM")
	v.Init()

}

func (v *VM) Init() {
	v.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	v.PC = 0x200
	v.SP = 0
	v.Timer = 0
	v.Sound = 0
	v.I = 0
	v.LoadFont()
	for i := 80; i < RAMSIZE; i++ {
		v.Memory[i] = 0
	}
	for i := 0; i < STACKSIZE; i++ {
		v.Stack[i] = 0
	}
	for i := 0; i < SCREENSIZE; i++ {
		v.Screen[i] = 0
	}
	for i := 0; i < 16; i++ {
		v.Keys[i] = 0
		v.Registers[i] = 0
	}

}

func (v *VM) LoadFont() {
	v.Memory = [RAMSIZE]uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0,
		0x20, 0x60, 0x20, 0x20, 0x70,
		0xF0, 0x10, 0xF0, 0x80, 0xF0,
		0xF0, 0x10, 0xF0, 0x10, 0xF0,
		0x90, 0x90, 0xF0, 0x10, 0x10,
		0xF0, 0x80, 0xF0, 0x10, 0xF0,
		0xF0, 0x80, 0xF0, 0x90, 0xF0,
		0xF0, 0x10, 0x20, 0x40, 0x40,
		0xF0, 0x90, 0xF0, 0x90, 0xF0,
		0xF0, 0x90, 0xF0, 0x10, 0xF0,
		0xF0, 0x90, 0xF0, 0x90, 0x90,
		0xE0, 0x90, 0xE0, 0x90, 0xE0,
		0xF0, 0x80, 0x80, 0x80, 0xF0,
		0xE0, 0x90, 0x90, 0x90, 0xE0,
		0xF0, 0x80, 0xF0, 0x80, 0xF0,
		0xF0, 0x80, 0xF0, 0x80, 0x80}
}

// Lod the file into memory starting at 0x200
func (v *VM) LoadProgram(program string) error {
	rom, err := os.ReadFile(program)
	if err != nil {
		return err
	}
	if v.debug {
		fmt.Println("Loading program:", program)
		fmt.Println("Program size:", len(rom))
	}
	for i := 0; i < len(rom); i++ {
		v.Memory[0x200+i] = rom[i]
	}
	return nil
}

// Fetch the next opcode from memory
func (v *VM) FetchOpCode() {

	v.OpCode = uint16(v.Memory[v.PC])<<8 | uint16(v.Memory[v.PC+1]) // 2 bytes opcode
	v.x = uint8((v.OpCode & 0x0F00) >> 8)                           // Decode Vx register
	v.y = uint8((v.OpCode & 0x00F0) >> 4)                           // Decode Vy register
	v.n = uint8(v.OpCode & 0x000F)                                  //  n is the last nibble of the opcode
	v.nn = uint8(v.OpCode & 0x00FF)                                 // nn is the last two bytes of the opcode
	v.nnn = uint16(v.OpCode & 0x0FFF)                               // nnn is the last three bytes of the opcode

	if v.debug {
		fmt.Printf("PC: %04X OpCode: %04X x: %X y: %X n: %X nn: %X nnn: %X ", v.PC, v.OpCode, v.x, v.y, v.n, v.nn, v.nnn)
	}
	v.PC += 2
}

// Execute the opcode
func (v *VM) ExecuteOpCode() {
	v.FetchOpCode()

	switch v.OpCode & 0xF000 {
	case 0x00E0:
		for i := 0; i < SCREENSIZE; i++ {
			v.Screen[i] = 0
		}
	case 0x00EE:
		v.SP--
		v.PC = v.Stack[v.SP]
	case 0x1000:
		v.PC = v.nnn
	case 0x2000:
		v.Stack[v.SP] = v.PC
		v.SP++
		v.PC = v.nnn
	case 0x3000:
		if v.Registers[v.x] == v.nn {
			v.PC += 2
		}
	case 0x4000:
		if v.Registers[v.x] != v.nn {
			v.PC += 2
		}
	case 0x5000:
		if v.Registers[v.x] == v.Registers[v.y] {
			v.PC += 2
		}
	case 0x6000:
		v.Registers[v.x] = v.nn
	case 0x7000:
		v.Registers[v.x] += v.nn
	case 0x8000:
		switch v.OpCode & 0x000F {
		case 0x0000:
			v.Registers[v.x] = v.Registers[v.y]
		case 0x0001:
			v.Registers[v.x] |= v.Registers[v.y]
		case 0x0002:
			v.Registers[v.x] &= v.Registers[v.y]
		case 0x0003:
			v.Registers[v.x] ^= v.Registers[v.y]
		case 0x0004:
			tx := v.Registers[v.x]
			ty := v.Registers[v.y]
			if int(tx)+int(ty) > 255 {
				v.Registers[0xF] = 1
			} else {
				v.Registers[0xF] = 0
			}
			v.Registers[v.x] += v.Registers[v.y]
		case 0x0005:
			tx := v.Registers[v.x]
			ty := v.Registers[v.y]
			if int(tx) > int(ty) {
				//v.StoreRegister(0xF, 1)
				v.Registers[0xF] = 1
			} else {
				v.Registers[0xF] = 0
			}
			v.Registers[v.x] -= v.Registers[v.y]
		case 0x0006:
			v.Registers[0xF] = v.Registers[v.x] & 0x1
			v.Registers[v.x] /= 2
		case 0x0007:
			if int(v.Registers[v.y]) > int(v.Registers[v.x]) {
				v.Registers[0xF] = 1
			} else {
				v.Registers[0xF] = 0
			}
			v.Registers[v.x] = v.Registers[v.y] - v.Registers[v.x]
		case 0x000E:
			v.Registers[0xF] = v.Registers[v.x] >> 7
			v.Registers[v.x] *= 2
		}
	case 0x9000:
		if v.Registers[v.x] != v.Registers[v.y] {
			v.PC += 2
		}
	case 0xA000:
		v.I = v.nnn
	case 0xB000:
		v.PC = v.nnn + uint16(v.Registers[0])
	case 0xC000:
		v.Registers[v.x] = uint8(v.rnd.Intn(256)) & v.nn
	case 0xD000:
		if v.debug {
			fmt.Print("Dxyn - DRW Vx, Vy, nibble")
		}
		yi := v.Registers[v.y]
		xi := v.Registers[v.x]
		v.Registers[0xF] = 0

		for k := 0; k < int(v.n); k++ {
			iv := v.I + uint16(k)
			if iv >= 4096 {
				fmt.Println("ERROR: v.I + k out of bounds", iv)
				//v.Reset()
			}
			q := v.Memory[iv]
			for j := 0; j < 8; j++ {
				b := (q >> (7 - j)) & 0x1
				if b == 1 {
					tindex := XYToIndex((uint8(int(xi)+j) % 64), (uint8(int(yi)+k) % 32))
					if tindex >= 2048 {
						fmt.Println("ERROR: tindex out of bounds", tindex, (uint8(int(xi)+j) % 64), (uint8(int(yi)+k) % 32))
						//v.Reset()
					}
					v.Screen[tindex] ^= 1
					if v.Screen[tindex] == 0 {
						v.Registers[0xF] = 1
					}

				}
			}
		}

	case 0xF000:
		switch v.OpCode & 0x00FF {
		case 0x0007:
			v.Registers[v.x] = v.Timer
		case 0x000A:
		case 0x0015:
			v.Timer = v.Registers[v.x]
		case 0x0018:
			v.Sound = v.Registers[v.x]
		case 0x001E:
			v.I += uint16(v.Registers[v.x])
		case 0x0029:
			v.I = uint16(v.Registers[v.x]) * 5
		case 0x0033:
			v.Memory[v.I] = v.Registers[v.x] / 100
			v.Memory[v.I+1] = (v.Registers[v.x] / 10) % 10
			v.Memory[v.I+2] = (v.Registers[v.x] % 100) % 10
		case 0x0055:
			for i := 0; i <= int(v.x); i++ {
				v.Memory[v.I+uint16(i)] = v.Registers[i]
			}
		case 0x0065:
			for i := 0; i <= int(v.x); i++ {
				v.Registers[i] = v.Memory[v.I+uint16(i)]
			}
		}
	}
	if v.debug {
		fmt.Println()
	}
}

// Convert x,y coordinates to screen index
func XYToIndex(x uint8, y uint8) int {
	return int(y)*SCREENWIDTH + int(x)
}

// Convert screen index to x,y coordinates
func IndexToXY(index int) (uint8, uint8) {
	return uint8(index % SCREENWIDTH), uint8(index / SCREENWIDTH)
}
