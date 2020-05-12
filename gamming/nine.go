package gamming

import (
	"github.com/hajimehoshi/ebiten"
	"image"
)

type Nine struct {
	Images              *ebiten.Image
	Alpha               float64
	R, G, B, Scale      float64
	Positions           [4][2]int
	x, y, width, height int
	scaleCenterWidth    float64
	scaleCenterHeight   float64
	targetPositions     [4][2]float64
}

func (n *Nine) SetColor(R, G, B float64) {
	n.R, n.G, n.B = R, G, B
}

func (n *Nine) SetPosition(x, y int) {
	n.x = x
	n.y = y
	n.SetSize(n.width, n.height)
}

func (n *Nine) SetSize(width, height int) {
	n.width = width
	n.height = height
	n.targetPositions[0][0] = float64(n.x)
	n.targetPositions[0][1] = float64(n.y)

	n.targetPositions[1][0] = float64(n.x) + n.Scale*float64(n.Positions[1][0])
	n.targetPositions[1][1] = float64(n.y) + n.Scale*float64(n.Positions[1][1])

	n.targetPositions[2][0] = float64(n.x+n.width) - n.Scale*float64(n.Positions[3][0]-n.Positions[2][0])
	n.targetPositions[2][1] = float64(n.y+n.height) - n.Scale*float64(n.Positions[3][1]-n.Positions[2][1])

	innerWidth := n.targetPositions[2][0] - n.targetPositions[1][0]
	innerHigh := n.targetPositions[2][1] - n.targetPositions[1][1]

	n.scaleCenterWidth = float64(innerWidth) / float64(n.Positions[2][0]-n.Positions[1][0])
	n.scaleCenterHeight = float64(innerHigh) / float64(n.Positions[2][1]-n.Positions[1][1])

}

func (n *Nine) Draw(screen *ebiten.Image) {

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.Scale, n.Scale)
	op.GeoM.Translate(n.targetPositions[0][0], n.targetPositions[0][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[0][0], n.Positions[0][1], n.Positions[1][0], n.Positions[1][1])).(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.scaleCenterWidth, n.Scale)
	op.GeoM.Translate(n.targetPositions[1][0], n.targetPositions[0][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[1][0], n.Positions[0][1], n.Positions[2][0], n.Positions[1][1])).(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.Scale, n.Scale)
	op.GeoM.Translate(n.targetPositions[2][0], n.targetPositions[0][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[2][0], n.Positions[0][1], n.Positions[3][0], n.Positions[1][1])).(*ebiten.Image), op)
	//-----
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.Scale, n.scaleCenterHeight)
	op.GeoM.Translate(n.targetPositions[0][0], n.targetPositions[1][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[0][0], n.Positions[1][1], n.Positions[1][0], n.Positions[2][1])).(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.scaleCenterWidth, n.scaleCenterHeight)
	op.GeoM.Translate(n.targetPositions[1][0], n.targetPositions[1][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[1][0], n.Positions[1][1], n.Positions[2][0], n.Positions[2][1])).(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.Scale, n.scaleCenterHeight)
	op.GeoM.Translate(n.targetPositions[2][0], n.targetPositions[1][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[2][0], n.Positions[1][1], n.Positions[3][0], n.Positions[2][1])).(*ebiten.Image), op)
	//----
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.Scale, n.Scale)
	op.GeoM.Translate(n.targetPositions[0][0], n.targetPositions[2][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[0][0], n.Positions[2][1], n.Positions[1][0], n.Positions[3][1])).(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.scaleCenterWidth, n.Scale)
	op.GeoM.Translate(n.targetPositions[1][0], n.targetPositions[2][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[1][0], n.Positions[2][1], n.Positions[2][0], n.Positions[3][1])).(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(n.Scale, n.Scale)
	op.GeoM.Translate(n.targetPositions[2][0], n.targetPositions[2][1])
	op.ColorM.Scale(n.R, n.G, n.B, n.Alpha)
	screen.DrawImage(n.Images.SubImage(image.Rect(n.Positions[2][0], n.Positions[2][1], n.Positions[3][0], n.Positions[3][1])).(*ebiten.Image), op)

}
