package gamelogic

import (
	"math"
)

func UnitVector(tx, ty, x, y float64) (ux, uy float64) {
	if tx == x && ty == y {
		return 0, 0
	}
	dx := tx - x
	dy := ty - y
	mag := math.Sqrt(dx*dx + dy*dy)
	ux = dx / mag
	uy = dy / mag
	return ux, uy
}

func MoveToTarget(tx, ty, x, y, speed float64) (ux, uy float64) {
	ux, uy = UnitVector(tx, ty, x, y)
	if ux == 0 && uy == 0 {
		return 0, 0
	}
	dx, dy := speed*ux, speed*uy
	nextX, nextY := dx+x, dy+y
	curDistSq := (tx-x)*(tx-x) + (ty-y)*(ty-y)
	nextDistSq := (tx-nextX)*(tx-nextX) + (ty-nextY)*(ty-nextY)

	if nextDistSq >= curDistSq {
		return tx - x, ty - y
	}
	return dx, dy
}
