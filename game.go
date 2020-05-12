package main

import (
	"bytes"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/tanema/gween"
	"github.com/zucenko/roader/client"
	"github.com/zucenko/roader/model"
	"github.com/zucenko/roaderclient/gamming"
	"golang.org/x/image/font"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	size = 40
)

func HexToF32(u uint32, id int) GameColor {
	b := float64(0xff&u) / 255
	g := float64(0xff&(u>>8)) / 255
	r := float64(0xff&(u>>16)) / 255
	return GameColor{r, g, b, id}
}

type GameColor struct {
	r  float64
	g  float64
	b  float64
	id int
}

var cols = 20
var rows = 13
var screenWidth = cols * size
var screenHeight = rows * size

var COLOR_NONE = HexToF32(0x000000, 0)
var COLOR_WHITE = HexToF32(0xffffff, 0)
var COLOR_DIAMOND = HexToF32(0xffe900, 0)
var COLOR_STONE = HexToF32(0x555555, 7)

var COLORS = []GameColor{
	HexToF32(0x1fc4ff, 1),
	HexToF32(0xff83f7, 2),
	HexToF32(0x68ff1f, 3),
}

// Tile represents an image.
type Tile struct {
	image          *ebiten.Image
	x              int
	y              int
	scaleX, scaleY float64
	color          GameColor
	alpha          float64
	width, height  int
}

func NewTile(image *ebiten.Image) *Tile {
	width, height := image.Size()
	return &Tile{
		image:  image,
		scaleX: 1,
		scaleY: 1,
		color:  COLOR_WHITE,
		width:  width,
		height: height}
}
func (t *Tile) SetSize(scale float64) {
	t.scaleX = scale
	t.scaleY = scale
	width, height := t.image.Size()
	t.width = int(float64(width) * scale)
	t.height = int(float64(height) * scale)

}

func (t *Tile) SetColor(gc GameColor) {
	t.color = gc
}

// Draw draws the sprite.
func (s *Tile) Draw(screen *ebiten.Image, dx, dy int, alpha float64, offset float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s.scaleX*.92, s.scaleY*.92) //s.width, s.height)
	op.GeoM.Translate(float64(s.x+dx), float64(s.y+dy)-offset)
	op.GeoM.Rotate(0)
	op.ColorM.Scale(s.color.r, s.color.g, s.color.b, alpha)
	screen.DrawImage(s.image, op)
}

func (s *Tile) DrawCentered(screen *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s.scaleX, s.scaleY)
	op.GeoM.Translate(float64(x-s.width/2), float64(y-s.height/2))
	op.ColorM.Scale(s.color.r, s.color.g, s.color.b, 1)
	screen.DrawImage(s.image, op)
}

// StrokeSource represents a input device to provide strokes.
type StrokeSource interface {
	Position() (int, int)
	IsJustReleased() bool
}

// MouseStrokeSource is a StrokeSource implementation of mouse.
type MouseStrokeSource struct{}

func (m *MouseStrokeSource) Position() (int, int) {
	return ebiten.CursorPosition()
}

func (m *MouseStrokeSource) IsJustReleased() bool {
	return inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
}

// TouchStrokeSource is a StrokeSource implementation of touch.
type TouchStrokeSource struct {
	ID int
}

func (t *TouchStrokeSource) Position() (int, int) {
	return ebiten.TouchPosition(t.ID)
}

func (t *TouchStrokeSource) IsJustReleased() bool {
	return inpututil.IsTouchJustReleased(t.ID)
}

// Stroke manages the current drag state by mouse.
type Stroke struct {
	source StrokeSource

	// initX and initY represents the position when dragging starts.
	initX int
	initY int

	// currentX and currentY represents the current position
	currentX int
	currentY int

	released bool
}

func NewStroke(source StrokeSource) *Stroke {
	cx, cy := source.Position()
	return &Stroke{
		source:   source,
		initX:    cx,
		initY:    cy,
		currentX: cx,
		currentY: cy,
	}
}

func (s *Stroke) Update() {
	if s.released {
		return
	}
	if s.source.IsJustReleased() {
		s.released = true
		return
	}
	x, y := s.source.Position()
	s.currentX = x
	s.currentY = y
}

func (s *Stroke) IsReleased() bool {
	return s.released
}

func (s *Stroke) Position() (int, int) {
	return s.currentX, s.currentY
}

func (s *Stroke) PositionDiff() (int, int) {
	dx := s.currentX - s.initX
	dy := s.currentY - s.initY
	return dx, dy
}

type GameState int

const (
	IDLE GameState = iota + 1
	COOLDOWN
	ACTING
	GAME_OVER
)

func (s GameState) Name() string {
	switch s {
	case IDLE:
		return "IDLE"
	case COOLDOWN:
		return "COOLDOWN"
	case ACTING:
		return "ACTING"
	case GAME_OVER:
		return "GAME_OVER"
	default:
		return fmt.Sprintf("N/A(%d)", s)
	}
}

type Play struct {
	State       GameState
	GameSession *client.GameSession
	strokes     map[*Stroke]struct{}
	Cols, Rows  int
	Tweens      map[*gween.Tween]gamming.Action
}

func (play *Play) move(dir int) {
	log.Printf(">> PRESSED %v", dir)
	play.GameSession.MessagesOut <- model.ClientMessage{Move: dir}
}

var play *Play
var imgDiamond, imgDiamondIn, imgPortal, imgPlayer, imgDot, imgDotSmall, imgKey *Tile
var bgImage, eImmF *ebiten.Image
var Font font.Face
var Line *gamming.Nine

var scale = 1.0
var wallWidth = 9

func init() {

	//Load()

	dat, err := ebitenutil.OpenFile("graphics/MiriamLibre-Bold.ttf")
	//dat, err := ioutil.ReadFile("Teko-Light.ttf")

	buf := new(bytes.Buffer)
	buf.ReadFrom(dat)

	tt, err := truetype.Parse(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	Font = truetype.NewFace(tt, &truetype.Options{
		Size:       50,
		DPI:        dpi,
		SubPixelsX: 100,
		Hinting:    font.HintingFull,
	})

	dot, _, err := ebitenutil.NewImageFromFile("graphics/circle.png", ebiten.FilterDefault)
	//imgF, _, err := image.Decode(bufio.NewReader(fileF))
	if err != nil {
		log.Fatal(err)
	}
	wallWidth = 8

	Line = &gamming.Nine{
		Images: dot,
		Alpha:  1,
		R:      1, G: 1, B: 1, Scale: .67,
		Positions: [4][2]int{{0, 0}, {5, 5}, {6, 6}, {11, 11}}}

	//fileF, err := os.Open("frame.png")
	imgDiamondE, _, err := ebitenutil.NewImageFromFile("graphics/diamond.png", ebiten.FilterLinear)
	//imgF, _, err := image.Decode(bufio.NewReader(fileF))
	if err != nil {
		log.Fatal(err)
	}
	imgDiamond = NewTile(imgDiamondE)
	imgDiamond.SetSize(scale)

	imgDiamondInE, _, err := ebitenutil.NewImageFromFile("graphics/diamond_in.png", ebiten.FilterLinear)
	//imgF, _, err := image.Decode(bufio.NewReader(fileF))
	if err != nil {
		log.Fatal(err)
	}
	imgDiamondIn = NewTile(imgDiamondInE)
	imgDiamondIn.SetSize(scale)

	imgKeyE, _, err := ebitenutil.NewImageFromFile("graphics/key.png", ebiten.FilterLinear)
	if err != nil {
		log.Fatal(err)
	}
	imgKey = NewTile(imgKeyE)
	imgKey.SetSize(scale)

	imgPortalE, _, err := ebitenutil.NewImageFromFile("graphics/portal.png", ebiten.FilterLinear)
	//imgF, _, err := image.Decode(bufio.NewReader(fileF))
	if err != nil {
		log.Fatal(err)
	}
	imgPortal = NewTile(imgPortalE)
	imgPortal.SetSize(scale)

	imgPlayerE, _, err := ebitenutil.NewImageFromFile("graphics/ring.png", ebiten.FilterLinear)
	if err != nil {
		log.Fatal(err)
	}
	imgPlayer = NewTile(imgPlayerE)
	imgPlayer.SetSize(scale)

	imgDotE, _, err := ebitenutil.NewImageFromFile("graphics/circle.png", ebiten.FilterLinear)
	if err != nil {
		log.Fatal(err)
	}
	imgDot = NewTile(imgDotE)
	imgDot.SetSize(scale)

	imgDotSmall = NewTile(imgDotE)
	imgDotSmall.SetSize(scale * .6)
	imgDotSmall.SetColor(COLOR_STONE)

	gs := client.NewGameSession()

	//go c.ServerMessageProcessingLoop()

	play = &Play{
		strokes:     map[*Stroke]struct{}{},
		Tweens:      make(map[*gween.Tween]gamming.Action),
		State:       IDLE,
		GameSession: gs,
		//ScoreLabel: prepareTextImage("00000"),
		//LevelLabel: prepareTextImage("1"),
		//score:      0,
	}

	//go gs.Loop()

	gs.Connect("i.glow.cz:8080")
}

func (play *Play) updateStroke(stroke *Stroke) {
	stroke.Update()
	xDif, yDif := stroke.PositionDiff()
	if math.Abs(float64(xDif)) > size/2 {
		stroke.released = true
	}

	if math.Abs(float64(yDif)) > size/2 {
		stroke.released = true
		// send event
	}

	if !stroke.IsReleased() {
		return
	}

	//-----------------------------------------------------

	//newX, _ := stroke.PositionDiff()
	//dx := float32(newX - 0)

}

var scoreImg *ebiten.Image
var keys, diamonds int

func prepareTextImage(keys int, diamonds int) *ebiten.Image {
	image, _ := ebiten.NewImage(220, 55, ebiten.FilterLinear)
	image.Fill(color.RGBA{0, 0, 0, 105})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, 1)
	op.GeoM.Translate(5, 15)
	image.DrawImage(imgKey.image, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, 1)
	op.GeoM.Translate(185, 15)
	image.DrawImage(imgDiamond.image, op)

	text.Draw(image, fmt.Sprintf("%d     %d", keys, diamonds), Font, 45, 45, color.White)
	return image
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func drawWall(screen *ebiten.Image, d int, topX int, topY int, lock bool, canUnlock bool, cell *model.Cell) {
	switch d {
	case 0:
		if lock {
			if canUnlock {
				imgDotSmall.color = COLOR_WHITE
			} else {
				imgDotSmall.color = COLOR_STONE
			}
			imgDotSmall.DrawCentered(screen, topX+cell.Col*size+size/2, topY+cell.Row*size-size/4)
			imgDotSmall.DrawCentered(screen, topX+cell.Col*size+size/2, topY+cell.Row*size)
			imgDotSmall.DrawCentered(screen, topX+cell.Col*size+size/2, topY+cell.Row*size+size/4)
		} else {
			Line.SetPosition(
				topX+cell.Col*size+size/2-wallWidth/2,
				topY+cell.Row*size-size/2-wallWidth/2)
			Line.SetSize(wallWidth, size+wallWidth)
			Line.Draw(screen)
		}
	case 1:
		if lock {
			if canUnlock {
				imgDotSmall.color = COLOR_WHITE
			} else {
				imgDotSmall.color = COLOR_STONE
			}
			imgDotSmall.DrawCentered(screen, topX+cell.Col*size-size/4, topY+cell.Row*size+size/2)
			imgDotSmall.DrawCentered(screen, topX+cell.Col*size, topY+cell.Row*size+size/2)
			imgDotSmall.DrawCentered(screen, topX+cell.Col*size+size/4, topY+cell.Row*size+size/2)
		} else {
			Line.SetPosition(
				topX+cell.Col*size-size/2-wallWidth/2,
				topY+cell.Row*size+size/2-wallWidth/2)
			Line.SetSize(size+wallWidth, wallWidth)
			Line.Draw(screen)
		}
	case 2:
		Line.SetPosition(topX+cell.Col*size-size/2-wallWidth/2, topY+cell.Row*size-size/2-wallWidth/2)
		Line.SetSize(wallWidth, size+wallWidth)
		Line.Draw(screen)
	case 3:
		Line.SetPosition(topX+cell.Col*size-size/2-wallWidth/2, topY+cell.Row*size-size/2-wallWidth/2)
		Line.SetSize(size+wallWidth, wallWidth)
		Line.Draw(screen)
	}
}

func colorForPlayer(playerKey int32) GameColor {
	switch playerKey {
	case 'A':
		return COLORS[0]
	case 'B':
		return COLORS[1]
	case 'C':
		return COLORS[2]
	default:
		return COLOR_STONE
	}
}

func (play *Play) update(screen *ebiten.Image) error {
	//log.Print("dddddd")

	play.GameSession.Loop()

	// tween
	for t, a := range play.Tweens {
		curr, finished := t.Update(0.02)
		if a.OnChange != nil {
			a.OnChange(curr)
		}
		if finished {
			for _, onFinish := range a.OnFinish {
				onFinish()
			}
			delete(play.Tweens, t)
		}
	}

	var pressed []ebiten.Key
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if ebiten.IsKeyPressed(k) {
			pressed = append(pressed, k)
		}
	}

	if repeatingKeyPressed(ebiten.KeyRight) {
		play.move(0)
	}
	if repeatingKeyPressed(ebiten.KeyDown) {
		play.move(1)
	}
	if repeatingKeyPressed(ebiten.KeyLeft) {
		play.move(2)
	}
	if repeatingKeyPressed(ebiten.KeyUp) {
		play.move(3)
	}
	if repeatingKeyPressed(ebiten.KeySpace) {
		play.move(4)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s := NewStroke(&MouseStrokeSource{})
		//s.SetDraggingObject(play.spriteAt(s.Position()))
		play.strokes[s] = struct{}{}
	}

	for _, id := range inpututil.JustPressedTouchIDs() {
		s := NewStroke(&TouchStrokeSource{id})
		//s.SetDraggingObject(play.spriteAt(s.Position()))
		play.strokes[s] = struct{}{}
	}

	for s := range play.strokes {
		play.updateStroke(s)
		if s.IsReleased() {
			delete(play.strokes, s)
		}
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	e := screen.Fill(color.RGBA{0, 0, 0, 255})

	if e != nil {
		log.Printf("%v", e)
	}

	topX := size
	topY := size

	if play.GameSession.Model != nil {

		if play.GameSession.Model.Players[play.GameSession.PlayerKey].Diamonds != diamonds ||
			play.GameSession.Model.Players[play.GameSession.PlayerKey].Keys != keys {
			diamonds = play.GameSession.Model.Players[play.GameSession.PlayerKey].Diamonds
			keys = play.GameSession.Model.Players[play.GameSession.PlayerKey].Keys
			scoreImg = prepareTextImage(keys, diamonds)
		}

		for c := 0; c < len(play.GameSession.Model.Matrix); c++ {
			for r := 0; r < len(play.GameSession.Model.Matrix[0]); r++ {
				cell := play.GameSession.Model.Matrix[c][r]
				Line.SetColor(COLOR_STONE.r, COLOR_STONE.g, COLOR_STONE.b)
				if c == 0 && cell.Paths[2].Wall {
					drawWall(screen, 2, topX, topY, false, false, cell)
				}
				if r == 0 && cell.Paths[3].Wall {
					drawWall(screen, 3, topX, topY, false, false, cell)
				}
				// NORMAL WALL
				if cell.Paths[0] != nil && cell.Paths[0].Wall {
					canUnlock := cell.Player != nil &&
						cell.Player.Keys > 0 ||
						cell.Paths[0].Target != nil &&
							cell.Paths[0].Target.Player != nil &&
							cell.Paths[0].Target.Player.Keys > 0
					drawWall(screen, 0, topX, topY, cell.Paths[0].Lock, canUnlock, cell)

				}
				// NORMAL WALL
				if cell.Paths[1] != nil && cell.Paths[1].Wall {
					canUnlock := cell.Player != nil &&
						cell.Player.Keys > 0 ||
						cell.Paths[1].Target != nil &&
							cell.Paths[1].Target.Player != nil &&
							cell.Paths[1].Target.Player.Keys > 0

					drawWall(screen, 1, topX, topY, cell.Paths[1].Lock, canUnlock, cell)
				}

				// PLAYER PATH SEGMENTS
				difStart := 0
				difLenBase := 0
				difLenFinal := 0
				CROSS_OFF := 7
				if cell.Crossing {
					difStart = CROSS_OFF
					difLenBase = CROSS_OFF
					difLenFinal = CROSS_OFF
				}
				if c < len(play.GameSession.Model.Matrix)-1 {
					path := cell.Paths[0]
					if path.Player != nil {
						if path.Target.Crossing {
							difLenFinal = difLenBase + CROSS_OFF
						} else {
							difLenFinal = difLenBase
						}
						color := colorForPlayer(path.Player.Id)
						Line.SetColor(color.r, color.g, color.b)
						Line.SetPosition(topX+c*size-wallWidth/2+difStart, topY+r*size-wallWidth/2)
						Line.SetSize(size+wallWidth-difLenFinal, wallWidth)
						Line.Draw(screen)
					}
				}

				if r < len(play.GameSession.Model.Matrix[0])-1 {
					path := cell.Paths[1]
					if path.Player != nil {
						if path.Target.Crossing {
							difLenFinal = difLenBase + CROSS_OFF
						} else {
							difLenFinal = difLenBase
						}
						color := colorForPlayer(path.Player.Id)
						Line.SetColor(color.r, color.g, color.b)
						Line.SetPosition(topX+c*size-wallWidth/2, topY+r*size-wallWidth/2+difStart)
						Line.SetSize(wallWidth, size+wallWidth-difLenFinal)
						Line.Draw(screen)
					}
				}
			}
		}

		for c := 0; c < len(play.GameSession.Model.Matrix); c++ {
			for r := 0; r < len(play.GameSession.Model.Matrix[0]); r++ {
				cell := play.GameSession.Model.Matrix[c][r]
				// DIAMOND
				if cell.Diamond {
					imgDiamondIn.SetColor(COLORS[0])
					imgDiamondIn.DrawCentered(screen, topX+c*size, topY+r*size)
					imgDiamond.SetColor(COLOR_DIAMOND)
					imgDiamond.DrawCentered(screen, topX+c*size, topY+r*size)
				}

				// KEY
				if cell.Key {
					imgKey.SetColor(COLOR_DIAMOND)
					imgKey.DrawCentered(screen, topX+c*size, topY+r*size)
				}

				// PORTAL
				if cell.Portal != nil {
					if cell.Player != nil {
						color := colorForPlayer(cell.Player.Id)
						imgPortal.SetColor(color)
					} else {
						clr := COLOR_WHITE
						for _, pressed := range cell.Paths {
							if pressed.Player != nil {
								clr = colorForPlayer(pressed.Player.Id)
							}
						}
						imgPortal.SetColor(clr)
					}
					imgPortal.DrawCentered(screen, topX+c*size, topY+r*size)
				}

				// PLAYER
				if cell.Player != nil {
					if cell.Portal == nil {
						imgPlayer.SetColor(colorForPlayer(cell.Player.Id))
						//log.Printf("[%d,%d] %v",c,r,cell.PlayerId.Id)
						imgPlayer.DrawCentered(screen, topX+c*size, topY+r*size)
					}
					imgDot.SetColor(colorForPlayer(cell.Player.Id))
					imgDot.DrawCentered(screen, topX+c*size, topY+r*size)
				}

			}
		}

	}
	if scoreImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1, 1)
		op.GeoM.Translate(100, 100)
		op.GeoM.Rotate(.1)
		screen.DrawImage(scoreImg, op)
	}
	/*
		//play.Score
		op := &ebiten.DrawImageOptions{}
		//op.GeoM.Scale(.35, .35) //s.width, s.height)
		op.GeoM.Translate(251, 10)
		op.GeoM.Rotate(0)
		//op.ColorM.Scale(0.05, .05, .05, 1)
		op = &ebiten.DrawImageOptions{}
		//op.GeoM.Scale(.35, .35) //s.width, s.height)
		op.GeoM.Scale(.7, 0.7)
		op.GeoM.Translate(252, 60)
		op.GeoM.Rotate(.2)
		//screen.DrawImage(play.LevelLabel, op)

		//ebitenutil.DebugPrintAt(screen, play.State.Name(), 300, 0)
	*/
	return nil
}

func main() {
	ebiten.SetRunnableInBackground(true)
	if err := ebiten.Run(play.update, screenWidth, screenHeight, 1, "Roader"); err != nil {
		log.Fatal(err)
	}
}
