package i8080Invaders

import (
	"github.com/veandco/go-sdl2/sdl"
)

var (
	WIDTH         = 224
	HEIGHT        = 256
	RED    uint32 = 0x0000FF
	CYAN   uint32 = 0xFFFF00
	GREEN  uint32 = 0x00FF00
	WHITE  uint32 = 0xFFFFFF
	BLACK  uint32 = 0x000000
)

type Screen struct {
	win *sdl.Window
	sur *sdl.Surface
	ren *sdl.Renderer
	tex *sdl.Texture
}

func NewScreen() *Screen {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	win := newWindow()
	ren := newRenderer(win)
	tex := newTexture(ren)
	sur := newSurface()
	screen := Screen{win: win, ren: ren, tex: tex, sur: sur}
	return &screen
}

func newWindow() *sdl.Window {
	win, err := sdl.CreateWindow("Space Invaders", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(WIDTH), int32(HEIGHT), sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		panic(err)
	}
	return win
}

func newRenderer(win *sdl.Window) *sdl.Renderer {
	ren, err := sdl.CreateRenderer(win, -1,
		sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	ren.SetLogicalSize(int32(WIDTH), int32(HEIGHT))
	return ren
}

func newTexture(ren *sdl.Renderer) *sdl.Texture {
	tex, err := ren.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32),
		sdl.TEXTUREACCESS_STREAMING, int32(WIDTH), int32(HEIGHT))
	if err != nil {
		panic(err)
	}
	return tex
}

func newSurface() *sdl.Surface {
	sur, err := sdl.CreateRGBSurface(0, int32(WIDTH), int32(WIDTH), 32, 0, 0, 0, 0)
	if err != nil {
		panic(err)
	}
	sur.SetRLE(true)
	return sur
}

func (s *Screen) Destroy() {
	s.tex.Destroy()
	s.ren.Destroy()
	s.win.Destroy()
	sdl.Quit()
}

func (s *Screen) Update() {
	s.ren.Copy(s.tex, nil, nil)
	s.ren.Present()
}

func (s *Screen) Draw(im *InvadersMachine) {
	for i := 0; i < (HEIGHT * WIDTH / 8); i++ {
		y0 := i * 8 / HEIGHT
		x0 := (i * 8) % HEIGHT
		curByte := im.cpu.GetMemory()[0x2400+i]

		for bit := uint8(0); bit < 8; bit++ {
			x := int32(x0 + int(bit))
			y := int32(y0)
			color := getColor(curByte, bit, x, y)
			tempX := x
			x = y
			y = -tempX + int32(HEIGHT) - 1
			s.drawPixel(x, y, color)
		}
	}
	s.updateTexture()
}

func (s *Screen) drawPixel(x int32, y int32, color uint32) {
	s.sur.FillRect(&sdl.Rect{X: x, Y: y, W: 1, H: 1}, color)
}

func (s *Screen) updateTexture() {
	pixels, _, err := s.tex.Lock(nil)
	if err != nil {
		panic(err)
	}
	copy(pixels, s.sur.Pixels())
	s.tex.Unlock()
}

func getColor(curByte uint8, curBit uint8, x int32, y int32) uint32 {
	pixelOn := ((curByte >> curBit) & 1) == 1
	if !pixelOn {
		return BLACK
	}
	if x < 16 && (y < 16 || y > 134) {
		return WHITE
	} else if x < 16 {
		return GREEN
	} else if x >= 16 && x <= 72 {
		return GREEN
	} else if x >= 192 && x < 224 {
		return RED
	} else {
		return CYAN
	}
}
