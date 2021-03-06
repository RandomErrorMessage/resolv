package main

import (
	"fmt"
	"math/rand"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SolarLune/resolv/resolv"
)

type WorldBounce struct {
	Space   *resolv.Space
	Squares []*Square
}

func (w *WorldBounce) Create() {

	var screenCellWidth = screenWidth / cell
	var screenCellHeight = screenHeight / cell

	// Just so nobody gets confused, yes, this isn't "true" fidelity because while I'm using floats for the speed variables,
	// I'm putting them into ints in the rectangle rather than having extra X and Y variables (just to make it easier to follow).

	w.Space = resolv.NewSpace()

	w.Space.Clear()
	w.Space.Add(resolv.NewRectangle(0, 0, screenWidth, cell))
	w.Space.Add(resolv.NewRectangle(0, cell, cell, screenHeight-cell))
	w.Space.Add(resolv.NewRectangle(screenWidth-cell, cell, cell, screenHeight-cell))
	w.Space.Add(resolv.NewRectangle(cell, screenHeight-cell, screenWidth-(cell*2), cell))

	for i := 0; i < 20; i++ {
		x := rand.Int31n(screenCellWidth - 2)
		y := rand.Int31n(screenCellHeight - 2)
		w.Space.Add(resolv.NewRectangle(cell+(x*cell), cell+(y*cell), cell*(1+rand.Int31n(16)), cell*(1+rand.Int31n(16))))
	}

	// Add the "solid" tag to all Shapes within the Space
	square := NewSquare(w.Space)
	w.Space.Add(square.Rect)
	w.Squares = append(w.Squares, square)
	w.Space.AddTags("solid")

}

func (w *WorldBounce) Update() {

	solids := w.Space.FilterByTags("solid")

	for _, Square := range w.Squares {

		Square.SpeedY += 0.25
		Square.BounceFrame *= .9

		if Square.SpeedY > float32(cell) {
			Square.SpeedY = float32(cell)
		} else if Square.SpeedY < -float32(cell) {
			Square.SpeedY = -float32(cell)
		}

		if Square.SpeedX > float32(cell) {
			Square.SpeedX = float32(cell)
		} else if Square.SpeedX < -float32(cell) {
			Square.SpeedX = -float32(cell)
		}

		// The additional teleporting check means that it won't resolve in a way that would cause it to move inordinately far (i.e.
		// teleporting). See the docs in resolv.go to see exactly what Teleporting is defined as.
		if res := solids.Resolve(Square.Rect, int32(Square.SpeedX), 0); res.Colliding() && !res.Teleporting {
			Square.Rect.X += res.ResolveX
			Square.SpeedX *= -1
			Square.BounceFrame = 1
		} else {
			Square.Rect.X += int32(Square.SpeedX)
		}

		if res := solids.Resolve(Square.Rect, 0, int32(Square.SpeedY)); res.Colliding() && !res.Teleporting {
			Square.Rect.Y += res.ResolveY
			Square.SpeedY *= -1
			// This makes the Squares able to rebound higher if they get a boost from another Square below~
			if Square.SpeedY < 0 && Square.SpeedY > -5 {
				Square.SpeedY = -5
			}
			Square.BounceFrame = 1
		} else {
			Square.Rect.Y += int32(Square.SpeedY)
		}

	}

	if rl.IsKeyDown(rl.KeyUp) {
		square := NewSquare(w.Space)
		w.Squares = append(w.Squares, square)
		w.Space.Add(square.Rect)
		fmt.Println(len(w.Squares), " Squares in the world now.")
	}

	if rl.IsKeyDown(rl.KeyDown) {

		squares := w.Space.FilterByTags("square")

		if squares.Length() > 0 {

			w.Space.Remove(squares.Get(0))

			for i, b := range w.Squares {

				if b.Rect == squares.Get(0) {
					w.Squares[i] = nil
					w.Squares = append(w.Squares[:i], w.Squares[i+1:]...)
				}

			}

			fmt.Println(len(w.Squares), " Squares in the world now.")

		}

	}

	if rl.IsKeyPressed(rl.KeyS) { // The ability to trigger solidity
		if !w.Squares[0].Rect.HasTags("solid") {
			w.Space.FilterByTags("square").AddTags("solid")
		} else {
			w.Space.FilterByTags("square").RemoveTags("solid")
		}
	}

}

func (w *WorldBounce) Draw() {

	for _, shape := range *w.Space {

		// Living on the edge~~~
		// We know that this Space just has Rectangles, so we'll just assume they are

		rect := shape.(*resolv.Rectangle)

		if !rect.HasTags("square") {

			rl.DrawRectangleLines(rect.X, rect.Y, rect.W, rect.H, rl.LightGray)

		} else {

			squareData := rect.GetData().(*Square)

			g := uint8(60) + uint8((255-60)*squareData.BounceFrame)

			color := rl.Color{g, g, g, 255}

			if rect.HasTags("solid") {
				color = rl.Color{60, g, 255, 255}
			}

			rl.DrawRectangleLines(squareData.Rect.X, squareData.Rect.Y, squareData.Rect.W, squareData.Rect.H, color)

		}

	}

	if drawHelpText {
		DrawText(32, 16,
			"-Bounce stress test-",
			"Press Up to spawn squares.",
			"Press Down to remove squares.",
			"Press 'S' to toggle solidity between the squares.",
			"Press 'R' to restart with a new",
			"layout.",
			strconv.Itoa(len(w.Squares))+" squares in the world",
		)
	}
}

func (w *WorldBounce) Destroy() {
	w.Squares = make([]*Square, 0)
	w.Space.Clear()
}
