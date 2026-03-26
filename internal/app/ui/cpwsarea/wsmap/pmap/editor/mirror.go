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

// TileMirrorSelected mirrors the currently selected tiles horizontally or vertically.
func (e *Editor) TileMirrorSelected(horizontal bool) {
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

	log.Printf("mirroring selected tiles, horizontal: %v, bounds: %dx%d", horizontal, W, H)

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

	// 2. Mirror those prefabs and calculate their new positions
	tilesToPaste := make(map[util.Point]dmmdata.Prefabs)
	for pt, prefabs := range oldPrefabs {
		relX := pt.X - minX
		relY := pt.Y - minY

		var newX, newY int
		if horizontal {
			newX = minX + (W - 1 - relX)
			newY = minY + relY
		} else {
			newX = minX + relX
			newY = minY + (H - 1 - relY)
		}
		newPos := util.Point{X: newX, Y: newY, Z: zLevel}

		mirroredPrefabs := make(dmmdata.Prefabs, 0, len(prefabs))
		for _, prefab := range prefabs {
			mirroredPrefabs = append(mirroredPrefabs, mirrorPrefab(prefab, horizontal))
		}
		tilesToPaste[newPos] = append(tilesToPaste[newPos], mirroredPrefabs...)
	}

	// Select grab tool and reset its selection
	toolSelect, ok := tools.SetSelected(tools.TNGrab).(*tools.ToolGrab)
	if ok {
		toolSelect.Reset()
		var tilesToSelect []util.Point
		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				pos := util.Point{X: x, Y: y, Z: zLevel}
				if e.Dmm().HasTile(pos) {
					tilesToSelect = append(tilesToSelect, pos)
				}
			}
		}

		toolSelect.PreSelectArea(tilesToSelect)

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

		toolSelect.SelectArea(tilesToSelect)
	}

	e.CommitChanges("Mirror Map Selection")
}

func mirrorPrefab(prefab *dmmprefab.Prefab, horizontal bool) *dmmprefab.Prefab {
	vars := prefab.Vars()
	newVars := vars

	// Mirror dir
	dirVal := vars.IntV("dir", 2)
	newDir := mirrorDir(dirVal, horizontal)
	if newDir != 2 { // 2 is SOUTH, default
		newVars = dmvars.Set(newVars, "dir", strconv.Itoa(newDir))
	} else {
		newVars = dmvars.Delete(newVars, "dir")
	}

	// Mirror pixel_x and pixel_y
	px := vars.IntV("pixel_x", 0)
	py := vars.IntV("pixel_y", 0)
	if horizontal && px != 0 {
		newVars = dmvars.Set(newVars, "pixel_x", strconv.Itoa(-px))
	} else if !horizontal && py != 0 {
		newVars = dmvars.Set(newVars, "pixel_y", strconv.Itoa(-py))
	}

	return dmmprefab.New(dmmprefab.IdNone, prefab.Path(), newVars)
}

func mirrorDir(dir int, horizontal bool) int {
	if horizontal {
		switch dir {
		case 4:
			return 8 // E -> W
		case 8:
			return 4 // W -> E
		case 5:
			return 9 // NE -> NW
		case 9:
			return 5 // NW -> NE
		case 6:
			return 10 // SE -> SW
		case 10:
			return 6 // SW -> SE
		}
	} else {
		switch dir {
		case 1:
			return 2 // N -> S
		case 2:
			return 1 // S -> N
		case 5:
			return 6 // NE -> SE
		case 6:
			return 5 // SE -> NE
		case 9:
			return 10 // NW -> SW
		case 10:
			return 9 // SW -> NW
		}
	}
	return dir
}
