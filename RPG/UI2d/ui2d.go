package ui2d

import (
	"bufio"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/t-RED-69/games-with-go/RPG/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var randSRC *rand.Rand

type ui struct {
	winWidht, winHeight int32
	renderer            *sdl.Renderer
	window              *sdl.Window
	tex                 *sdl.Texture
	zoom                int32
	centerX, centerY    int32
	textureAtlas        *[]SpriteTexture
	MiniAtlas           map[rune][]*SpriteTexture
	mouse               MouseState
	keyBoard            []KeyStates
	r                   *rand.Rand
	levelChan           chan *game.Level
	inputChan           chan *game.Input
	fontSmall           *ttf.Font
	fontMed             *ttf.Font
	fontLarg            *ttf.Font
	//
	str2TexSmll map[string]*sdl.Texture
	str2TexMed  map[string]*sdl.Texture
	str2TexLarg map[string]*sdl.Texture
	//
	eventBackGr  *sdl.Texture
	plyrDirnMrkr *sdl.Texture
}

const (
	DoorOpnINT int = iota
	FootstpsINT
	EnmyHitINT
	PlyrHitINT
)

var maxEventBGRWid int32

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.winWidht, ui.winHeight = 1280, 720
	ui.zoom = 3
	ui.str2TexSmll = make(map[string]*sdl.Texture)
	ui.str2TexMed = make(map[string]*sdl.Texture)
	ui.str2TexLarg = make(map[string]*sdl.Texture)
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.r = rand.New(rand.NewSource(1))
	window, err := sdl.CreateWindow("RPG !!", int32(1366/2-ui.winWidht/2), int32(766/2-ui.winHeight/2), int32(ui.winWidht), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window
	//defer window.Destroy() //defer executes this statement after reaching the end of function/finishing the execution of funtion
	//and we dont wanna destroy it

	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas = ui.SpriteOpener("UI2d/assets/tiles.png", 32, 32, 6042)
	ui.MiniAtlas = ui.idexAssignerToAtlas()

	ui.keyBoard = make([]KeyStates, len(sdl.GetKeyboardState()))
	ui.mouse.ProcessMouse()
	ProcessKeys(&ui.keyBoard)
	if ui.fontSmall, err = ttf.OpenFont("UI2d/assets/Kingthings_Foundation.ttf", 18); err != nil {
		fmt.Println("cannot open font: UI2d/assets/Kingthings_Foundation.ttf")
	}
	if ui.fontMed, err = ttf.OpenFont("UI2d/assets/Kingthings_Foundation.ttf", 32); err != nil {
		fmt.Println("cannot open font: UI2d/assets/Kingthings_Foundation.ttf")
	}
	if ui.fontLarg, err = ttf.OpenFont("UI2d/assets/Kingthings_Foundation.ttf", 48); err != nil {
		fmt.Println("cannot open font: UI2d/assets/Kingthings_Foundation.ttf")
	}
	ui.eventBackGr = ui.getSinglePixTex(sdl.Color{0, 0, 0, 100})
	ui.eventBackGr.SetBlendMode(sdl.BLENDMODE_BLEND)
	maxEventBGRWid = int32(float32(ui.winWidht) * 0.25)

	ui.plyrDirnMrkr = ui.getSinglePixTex(sdl.Color{220, 0, 0, 100})
	ui.eventBackGr.SetBlendMode(sdl.BLENDMODE_BLEND)

	return ui
}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

func (ui *ui) stringToTexture(s string, clor sdl.Color, size FontSize) *sdl.Texture {
	var tex *sdl.Texture
	var font *ttf.Font
	switch size {
	case FontSmall:
		tex, exist := ui.str2TexSmll[s]
		if exist {
			return tex
		}
		font = ui.fontSmall
	case FontMedium:
		tex, exist := ui.str2TexMed[s]
		if exist {
			return tex
		}
		font = ui.fontMed
	case FontLarge:
		tex, exist := ui.str2TexLarg[s]
		if exist {
			return tex
		}
		font = ui.fontLarg
	}
	fontSurface, err := font.RenderUTF8Blended(s, clor)
	if err != nil {
		panic(err)
	}
	tex, err = ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}
	switch size {
	case FontSmall:
		ui.str2TexSmll[s] = tex
	case FontMedium:
		ui.str2TexMed[s] = tex
	case FontLarge:
		ui.str2TexLarg[s] = tex
	}
	return tex
}

type MouseState struct {
	Left, Right        bool
	ChangedL, ChangedR bool
	X, Y               int32
}
type KeyStates struct {
	IsDown  bool
	Changed bool
}

func (m *MouseState) ProcessMouse() {
	x, y, mouse := sdl.GetMouseState()
	m.X, m.Y = x, y
	currL := (mouse&sdl.ButtonLMask() == 1)
	currR := (mouse&sdl.ButtonRMask() == 4)
	if m.Left != currL {
		m.ChangedL = true
	} else {
		m.ChangedL = false
	}
	if m.Right != currR {
		m.ChangedR = true
	} else {
		m.ChangedR = false
	}
	m.Left = currL
	m.Right = currR
}
func ProcessKeys(kb *[]KeyStates) {
	keystrokes := sdl.GetKeyboardState()
	for i := range *kb {
		if (*kb)[i].IsDown != (keystrokes[i] != 0) {
			(*kb)[i].Changed = true
		} else {
			(*kb)[i].Changed = false
		}
		(*kb)[i].IsDown = (keystrokes[i] != 0)
	}
}

//SpriteTexture cantains sprite's enum name,texture,default length and breadth for image
type SpriteTexture struct {
	symbol   rune
	varCount int
	index    int
	tex      *sdl.Texture
	len, bth int32
}

//SpriteOpener to load specified number of sprite textures
func (ui *ui) SpriteOpener(str string, lenPerSprite, widPerSprite int32, noOfSprites int) *[]SpriteTexture {
	inFile, err := os.Open(str)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	img, err := png.Decode(inFile)
	if err != nil {
		panic(err)
	}

	noOfColumn := int32(img.Bounds().Max.X / int(lenPerSprite))
	noOfRow := int32(int(float32(noOfSprites)/float32(noOfColumn)) + 1)
	var index int
	var r, g, b, a uint32
	spriteArray := make([]SpriteTexture, noOfSprites)
	var tex *sdl.Texture
	var i, j, x, y int32
	var counter, counter2 int
	for i = 0; i < noOfRow; i++ {
		for j = 0; j < noOfColumn; j++ {
			counter2++
			pixels := make([]byte, lenPerSprite*widPerSprite*4)
			index = 0
			for y = widPerSprite * i; y < widPerSprite*(i+1); y++ {
				for x = lenPerSprite * j; x < lenPerSprite*(j+1); x++ {
					r, g, b, a = img.At(int(x), int(y)).RGBA()
					pixels[index] = byte(r / 256)
					index++
					pixels[index] = byte(g / 256)
					index++
					pixels[index] = byte(b / 256)
					index++
					pixels[index] = byte(a / 256)
					index++
				}
			}
			tex, err = ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, lenPerSprite, widPerSprite)
			if err != nil {
				panic(err)
			}
			tex.Update(nil, pixels, int(lenPerSprite)*4)
			err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
			if err != nil {
				panic(err)
			}
			if (i*noOfColumn + j) < int32(noOfSprites) {
				spriteArray[i*noOfColumn+j] = SpriteTexture{' ', 0, 0, tex, lenPerSprite, widPerSprite}
				counter++
			} else {
				break
			}
		}
	}
	return &spriteArray
}
func (ui *ui) getSinglePixTex(color sdl.Color) *sdl.Texture {
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixS := make([]byte, 4)
	pixS[0] = color.R
	pixS[1] = color.G
	pixS[2] = color.B
	pixS[3] = color.A
	tex.Update(nil, pixS, 4)
	return tex
}
func init() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ttf.Init()
	if err != nil {
		fmt.Println(err)
		return
	}
}

//Draw to draw over screen
func (ui *ui) Draw(level *game.Level) {
	if (level.Player.X*ui.zoom - ui.centerX) > (ui.winWidht/2 + 64*ui.zoom) {
		ui.centerX += 3 * ui.zoom
	} else if (level.Player.X*ui.zoom - ui.centerX) < (ui.winWidht/2 - 64*ui.zoom) {
		ui.centerX -= 3 * ui.zoom
	} else if (level.Player.X*ui.zoom - ui.centerX) > (ui.winWidht / 2) {
		ui.centerX++
	} else if (level.Player.X*ui.zoom - ui.centerX) < (ui.winWidht / 2) {
		ui.centerX--
	}
	if (level.Player.Y*ui.zoom - ui.centerY) > (ui.winHeight/2 + 55*ui.zoom) {
		ui.centerY += 3 * ui.zoom
	} else if (level.Player.Y*ui.zoom - ui.centerY) < (ui.winHeight/2 - 55*ui.zoom) {
		ui.centerY -= 3 * ui.zoom
	} else if (level.Player.Y*ui.zoom - ui.centerY) > (ui.winHeight / 2) {
		ui.centerY++
	} else if (level.Player.Y*ui.zoom - ui.centerY) < (ui.winHeight / 2) {
		ui.centerY--
	}

	ui.renderer.Clear()
	ui.r.Seed(1)
	for y, row := range level.Map {
		var r int
		for x, tile := range row {
			if tile.Rune == 0 || tile.Rune == 32 { //0 is space//32 is \t
				continue
			}
			dstRect := sdl.Rect{int32(x*32)*ui.zoom - ui.centerX, int32(y*32)*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom}

			r = ui.r.Intn(ui.MiniAtlas[tile.Rune][0].varCount)
			//if level.Debug[game.Pos{int32(x), int32(y)}] {
			//	ui.MiniAtlas[tile.Rune][r].tex.SetColorMod(128, 0, 0)
			//} else {
			//	ui.MiniAtlas[tile.Rune][r].tex.SetColorMod(255, 255, 255)
			//}
			switch tile.Rune {
			case game.OpenDoor, game.StairUp:
				//if tile.Visible {
				ui.MiniAtlas['.'][0].tex.SetColorMod(255, 255, 255)
				//} else {
				//	ui.MiniAtlas['.'][0].tex.SetColorMod(100, 100, 100)
				//}
				if tile.Seen {
					ui.renderer.Copy(ui.MiniAtlas['.'][0].tex, nil, &dstRect)
				}
			}
			if tile.Visible {
				ui.MiniAtlas[tile.Rune][r].tex.SetColorMod(255, 255, 255)
			} else {
				//fmt.Println(r)
				ui.MiniAtlas[tile.Rune][r].tex.SetColorMod(100, 100, 100)
			}
			if tile.Seen {
				ui.renderer.Copy(ui.MiniAtlas[tile.Rune][r].tex, nil, &dstRect)
			}
		}
	}
	switch game.LookDirn {
	case game.Up:
		ui.renderer.Copy(ui.plyrDirnMrkr, nil, &sdl.Rect{(level.Player.X+14)*ui.zoom - ui.centerX, (level.Player.Y-7)*ui.zoom - ui.centerY, 4 * ui.zoom, 8 * ui.zoom})
	case game.Down:
		ui.renderer.Copy(ui.plyrDirnMrkr, nil, &sdl.Rect{(level.Player.X+14)*ui.zoom - ui.centerX, (level.Player.Y+30)*ui.zoom - ui.centerY, 4 * ui.zoom, 8 * ui.zoom})
	case game.Left:
		ui.renderer.Copy(ui.plyrDirnMrkr, nil, &sdl.Rect{(level.Player.X-4)*ui.zoom - ui.centerX, (level.Player.Y+14)*ui.zoom - ui.centerY, 8 * ui.zoom, 4 * ui.zoom})
	case game.Right:
		ui.renderer.Copy(ui.plyrDirnMrkr, nil, &sdl.Rect{(level.Player.X+30)*ui.zoom - ui.centerX, (level.Player.Y+14)*ui.zoom - ui.centerY, 8 * ui.zoom, 4 * ui.zoom})
	}

	for _, t := range level.Items {
		dstRect := sdl.Rect{int32(t.X)*ui.zoom - ui.centerX, int32(t.Y)*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom}
		ui.renderer.Copy(ui.MiniAtlas[t.Symbol][0].tex, nil, &dstRect)
	}

	ui.renderer.Copy(ui.MiniAtlas[level.Player.Symbol][0].tex, nil, &sdl.Rect{level.Player.X*ui.zoom - ui.centerX, level.Player.Y*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom})

	for _, j := range level.Monsters {
		dstRect := sdl.Rect{int32(j.X)*ui.zoom - ui.centerX, int32(j.Y)*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom}
		if !level.Map[j.Pos.Div32().Y][j.Pos.Div32().X].Visible {
			continue
		}
		if j.Debug2 {
			ui.MiniAtlas[j.Symbol][0].tex.SetColorMod(160, 160, 255)
		} else if j.Debug {
			ui.MiniAtlas[j.Symbol][0].tex.SetColorMod(255, 160, 160)
		} else {
			ui.MiniAtlas[j.Symbol][0].tex.SetColorMod(255, 255, 255)
		}
		ui.renderer.Copy(ui.MiniAtlas[j.Symbol][0].tex, nil, &dstRect)
	}
	//
	ui.renderer.Copy(ui.eventBackGr, nil, &sdl.Rect{5, int32(float32(ui.winHeight)*0.67 - 5), maxEventBGRWid + 8, int32(float32(ui.winWidht) * 0.25)})
	for i, j := len(level.Events)-1, 0; i >= 0; i, j = i-1, j+1 {
		tex := ui.stringToTexture(level.Events[i], sdl.Color{255, 20, 20, 0}, FontSmall)
		_, _, w, h, _ := tex.Query()
		ui.renderer.Copy(tex, nil, &sdl.Rect{10, int32(float32(ui.winHeight)*0.67 + float32(j*22)), w, h})
		if w > maxEventBGRWid {
			maxEventBGRWid = w
		}
	}
	k := len(level.Player.Items) - 1
	for _, j := range level.Player.Items {
		ui.renderer.Copy(ui.MiniAtlas[j.Symbol][0].tex, nil, &sdl.Rect{ui.winWidht - 64 - int32(k*48), 16, 48, 48})
		k--
	}
	var i int32
	for _, j := range level.Items {
		if level.Player.Pos.Div32() == j.Pos.Div32() {
			ui.renderer.Copy(ui.MiniAtlas[j.Symbol][0].tex, nil, &sdl.Rect{ui.winWidht - 64 - i*48, ui.winHeight - 64, 48, 48})
			i++
		}
	}
	//
	ui.renderer.Present()
	sdl.Delay(9)
}

func (ui *ui) idexAssignerToAtlas() map[rune][]*SpriteTexture {
	file, err := os.Open("UI2d/assets/tileSymbol-Index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	newAtlas := make(map[rune][]*SpriteTexture)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := rune(line[0])
		xyv := line[1:]
		splitXYV := strings.Split(xyv, ",")
		x, err := strconv.ParseInt(splitXYV[0], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(splitXYV[1], 10, 64)
		if err != nil {
			panic(err)
		}
		v, err := strconv.ParseInt(splitXYV[2], 10, 64)
		if err != nil {
			panic(err)
		}
		newAtlas[tileRune] = make([]*SpriteTexture, v)
		fmt.Printf("%c ", tileRune)
		var z int64
		for z = 0; z < v; z++ {
			(*ui.textureAtlas)[y*64+(x+z)].symbol = tileRune
			(*ui.textureAtlas)[y*64+(x+z)].varCount = int(v)
			(*ui.textureAtlas)[y*64+(x+z)].index = int(z)
			fmt.Printf(" %d", z)
			newAtlas[tileRune][z] = &(*ui.textureAtlas)[y*64+(x+z)]
		}
		fmt.Println()
	}
	return newAtlas
}
func (ui *ui) Run() {
	var lvle *game.Level
	var drawn bool
	for {
		ui.mouse.ProcessMouse()
		ProcessKeys(&ui.keyBoard)
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChannel: ui.levelChan}
				}
			}
		}
		select {
		case newLevel, ok := <-ui.levelChan:
			if drawn = ok; ok {
				if lvle != newLevel { //recalculating after level change
					lvle = newLevel
					ui.centerX = (lvle.Player.X*ui.zoom - ui.winWidht/2)
					ui.centerY = (lvle.Player.Y*ui.zoom - ui.winHeight/2)
				}
				ui.Draw(lvle)
			}
		default:
			drawn = false
		}
		//fmt.Println(drawn, ui.centerX != (lvle.Player.X*ui.zoom-ui.winWidht/2) || ui.centerY != (lvle.Player.Y*ui.zoom-ui.winHeight/2))
		if !drawn && (ui.centerX != (lvle.Player.X*ui.zoom-ui.winWidht/2) || ui.centerY != (lvle.Player.Y*ui.zoom-ui.winHeight/2)) {
			ui.Draw(lvle)
		}

		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			input := &game.Input{Typ: game.Blank}
			if ui.keyBoard[sdl.SCANCODE_DOWN].IsDown {
				input = &game.Input{Typ: game.Down}
			} else if ui.keyBoard[sdl.SCANCODE_UP].IsDown {
				input = &game.Input{Typ: game.Up}
			} else if ui.keyBoard[sdl.SCANCODE_LEFT].IsDown {
				input = &game.Input{Typ: game.Left}
			} else if ui.keyBoard[sdl.SCANCODE_RIGHT].IsDown {
				input = &game.Input{Typ: game.Right}
			} else if ui.keyBoard[sdl.SCANCODE_O].Changed && ui.keyBoard[sdl.SCANCODE_O].IsDown {
				input = &game.Input{Typ: game.Open}
			} else if ui.keyBoard[sdl.SCANCODE_SPACE].IsDown {
				input = &game.Input{Typ: game.EmptySpace}
			} else if ui.keyBoard[sdl.SCANCODE_S].Changed && ui.keyBoard[sdl.SCANCODE_S].IsDown {
				fmt.Println("search")
				input = &game.Input{Typ: game.Search}
			} else if ui.keyBoard[sdl.SCANCODE_P].IsDown {
				fmt.Println("pick")
				input = &game.Input{Typ: game.Pickup}
			} else if ui.keyBoard[sdl.SCANCODE_KP_PLUS].Changed && ui.keyBoard[sdl.SCANCODE_KP_PLUS].IsDown {
				ui.zoom++
				ui.centerX = (lvle.Player.X*ui.zoom - ui.winWidht/2)
				ui.centerY = (lvle.Player.Y*ui.zoom - ui.winHeight/2)
				ui.Draw(lvle)
			} else if ui.keyBoard[sdl.SCANCODE_KP_MINUS].Changed && ui.keyBoard[sdl.SCANCODE_KP_MINUS].IsDown {
				ui.zoom--
				ui.centerX = (lvle.Player.X*ui.zoom - ui.winWidht/2)
				ui.centerY = (lvle.Player.Y*ui.zoom - ui.winHeight/2)
				ui.Draw(lvle)
			}
			if input.Typ != game.Blank {
				ui.inputChan <- input
			}
		}
		sdl.Delay(10)
	}
}

/*
func Play(Sound int, str string) {
	var snd *mix.Chunk
	switch Sound {
	case DoorOpnINT:
		snd = uI.SFX.DoorOpen
	case FootstpsINT:
		PlayMrand(uI.SFX.Footsteps, FootstpsINT)
		return
	case EnmyHitINT:
		snd = uI.SFX.EnemyHit[str]
	case PlyrHitINT:
		snd = uI.SFX.PlyrHit
	}
	snd.Play(Sound, 0)
}
func HaltSounds(channel int) {
	mix.HaltChannel(channel)
}
func PlayMrand(snd []*mix.Chunk, channel int) {
	//rand.Rand.Seed(time.Now().UnixNano())
	randSRC.Seed(time.Now().UnixNano())
	ra := randSRC.Intn(6)
	snd[ra].Play(channel, 0)
}

/*

*/
