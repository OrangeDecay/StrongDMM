package editor

import (
	"sdmm/internal/app/ui/cpwsarea/wsmap/tools"
	"sdmm/internal/dmapi/dmmap/dmmdata"
	"sdmm/internal/dmapi/dmmap/dmmdata/dmmprefab"
	"sdmm/internal/dmapi/dmvars"
	"sdmm/internal/util"
	"strconv"

	"github.com/rs/zerolog/log"
)

// TileRotateSelected rotates the currently selected tiles by 90 degrees.
func (e *Editor) TileRotateSelected(clockwise bool) {
	selectedTiles := tools.SelectedTiles()
	if len(selectedTiles) == 0 {
		return
	}

	// Calculate bounds
	minX, minY, maxX, maxY := selectedTiles[0].X, selectedTiles[0].Y, selectedTiles[0].X, selectedTiles[0].Y
	zLevel := selectedTiles[0].Z

	for _, pt := range selectedTiles {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	W := maxX - minX + 1
	H := maxY - minY + 1

	log.Printf("rotating selected tiles, clockwise: %v, bounds: %dx%d", clockwise, W, H)

	// 1. For each coordinate in selection, collect visible prefabs and remove them
	var oldPrefabs = make(map[util.Point]dmmdata.Prefabs)
	for _, pt := range selectedTiles {
		if !e.Dmm().HasTile(pt) {
			continue
		}
		tile := e.Dmm().GetTile(pt)
		var visible dmmdata.Prefabs
		for _, instance := range tile.Instances() {
			if e.app.PathsFilter().IsVisiblePath(instance.Prefab().Path()) {
				visible = append(visible, instance.Prefab())
			}
		}
		if len(visible) > 0 {
			oldPrefabs[pt] = visible
			e.tileDelete(tile)
			tile.InstancesRegenerate()
		}
	}

	// 2. Rotate those prefabs and calculate their new positions
	tilesToPaste := make(map[util.Point]dmmdata.Prefabs)
	for pt, prefabs := range oldPrefabs {
		relX := pt.X - minX
		relY := pt.Y - minY

		var newX, newY int
		if clockwise {
			newX = minX + relY
			newY = minY + (W - 1 - relX)
		} else {
			newX = minX + (H - 1 - relY)
			newY = minY + relX
		}
		newPos := util.Point{X: newX, Y: newY, Z: zLevel}

		rotatedPrefabs := make(dmmdata.Prefabs, 0, len(prefabs))
		for _, prefab := range prefabs {
			rotatedPrefabs = append(rotatedPrefabs, rotatePrefab(prefab, clockwise))
		}
		tilesToPaste[newPos] = append(tilesToPaste[newPos], rotatedPrefabs...)
	}

	// 3. Paste them into their new locations
	for pos, prefabs := range tilesToPaste {
		if !e.Dmm().HasTile(pos) {
			continue
		}
		tile := e.Dmm().GetTile(pos)
		for _, prefab := range prefabs {
			tile.InstancesAdd(prefab)
		}
		tile.InstancesRegenerate()
	}

	// Select grab tool and reset its selection to match new bounds
	toolSelect, ok := tools.SetSelected(tools.TNGrab).(*tools.ToolGrab)
	if ok {
		toolSelect.Reset()
		var tilesToSelect []util.Point
		for x := minX; x <= minX+H-1; x++ {
			for y := minY; y <= minY+W-1; y++ {
				pos := util.Point{X: x, Y: y, Z: zLevel}
				if e.Dmm().HasTile(pos) {
					tilesToSelect = append(tilesToSelect, pos)
				}
			}
		}
		toolSelect.SelectArea(tilesToSelect)
	}

	e.CommitChanges("Rotate Map Selection")
}

func rotatePrefab(prefab *dmmprefab.Prefab, clockwise bool) *dmmprefab.Prefab {
	vars := prefab.Vars()
	newVars := vars

	// Rotate dir
	dirVal := vars.IntV("dir", 2)
	newDir := rotateDir(dirVal, clockwise)
	if newDir != 2 { // 2 is SOUTH, default
		newVars = dmvars.Set(newVars, "dir", strconv.Itoa(newDir))
	} else {
		newVars = dmvars.Delete(newVars, "dir")
	}

	// Rotate pixel_x and pixel_y
	px := vars.IntV("pixel_x", 0)
	py := vars.IntV("pixel_y", 0)
	if px != 0 || py != 0 {
		var newPx, newPy int
		if clockwise {
			newPx = py
			newPy = -px
		} else {
			newPx = -py
			newPy = px
		}

		if newPx != 0 {
			newVars = dmvars.Set(newVars, "pixel_x", strconv.Itoa(newPx))
		} else {
			newVars = dmvars.Delete(newVars, "pixel_x")
		}
		if newPy != 0 {
			newVars = dmvars.Set(newVars, "pixel_y", strconv.Itoa(newPy))
		} else {
			newVars = dmvars.Delete(newVars, "pixel_y")
		}
	}

	return dmmprefab.New(dmmprefab.IdNone, prefab.Path(), newVars)
}

func rotateDir(dir int, clockwise bool) int {
	if clockwise {
		switch dir {
		case 1:
			return 4 // N -> E
		case 2:
			return 8 // S -> W
		case 4:
			return 2 // E -> S
		case 8:
			return 1 // W -> N
		case 5:
			return 6 // NE -> SE
		case 6:
			return 10 // SE -> SW
		case 9:
			return 5 // NW -> NE
		case 10:
			return 9 // SW -> NW
		}
	} else {
		switch dir {
		case 1:
			return 8 // N -> W
		case 2:
			return 4 // S -> E
		case 4:
			return 1 // E -> N
		case 8:
			return 2 // W -> S
		case 5:
			return 9 // NE -> NW
		case 6:
			return 5 // SE -> NE
		case 9:
			return 10 // NW -> SW
		case 10:
			return 6 // SW -> SE
		}
	}
	return dir
}
