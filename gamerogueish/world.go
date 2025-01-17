package gamerogueish

import (
	"math/rand"
)

// World represents a game world.
type World struct {
	Cells    [][]byte  // 2D array of world cells
	Width    int       // width of the world in cells
	Height   int       // height of the world in cells
	Entities []*Entity // entities in the world (creatures)
}

// NewWorld returns a new world with the given width and height.
func NewWorld(width, height int) *World {
	w := &World{
		Width:  width,
		Height: height,
	}
	w.Cells = make([][]byte, height)
	for y := range w.Cells {
		w.Cells[y] = make([]byte, width)
	}
	return w
}

// IsSolid checks if a tile is solid (tile content is not a space ' ' character).
func (w *World) IsSolid(x int, y int) bool {
	return w.Cells[y][x] != ' '
}

// CanMoveTo checks if a tile is solid and if it is not occupied by an entity.
func (w *World) CanMoveTo(x, y int) bool {
	// TODO: Bounds check.
	return w.Cells[y][x] == ' '
}

// Fill all cells with the given tile.
func (w *World) Fill(c byte) {
	for y := range w.Cells {
		for x := range w.Cells[y] {
			w.Cells[y][x] = c
		}
	}
}

// InBounds returns true if the given position is within the world bounds.
func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

// CarveRoom sets all tiles occupied by the room to ' '.
func (w *World) CarveRoom(room *Room) {
	for y := room.Y; y < room.Y+room.H; y++ {
		for x := room.X; x < room.X+room.W; x++ {
			w.Cells[y][x] = ' '
		}
	}
}

// Cardinal directions.
const (
	DirNorth = 0
	DirEast  = 1
	DirSouth = 2
	DirWest  = 3
)

// GenWorldSimpleDungeon generates a simple random-walk-ish dungeon.
// - A starting room is placed in the center of the world.
// - Rooms are then placed in random directions neighboring a randomly selectd room.
// - Rooms are not placed if they would overlap with an existing room.
func GenWorldSimpleDungeon(width, height int, seed int64) *World {
	const (
		attempts    = 200
		maxRooms    = 100
		minRoomSize = 4
		maxRoomSize = 20
	)
	w := NewWorld(width, height)
	w.Fill('#')

	ssrc := rand.NewSource(seed)
	rng := rand.New(ssrc)

	// Start with a single room in the middle of the map.
	rl := randInt(rng, minRoomSize, maxRoomSize)
	rw := randInt(rng, minRoomSize, maxRoomSize)
	rooms := []*Room{{
		X: (width / 2) - rw/2,
		Y: (height / 2) - rl/2,
		W: rw,
		H: rl,
	}}

	// Carve out the starting room.
	w.CarveRoom(rooms[0])

	// Place rooms until we run out of attempts or reach the max room count.
	for i := 0; i < attempts; i++ {
		// Pick a room and place a neighboring room.
		room := rooms[rng.Intn(len(rooms))]
		// Pick a random direction.
		dir := rng.Intn(4)
		// Pick a random length and width.
		rl := randInt(rng, minRoomSize, maxRoomSize)
		rw := randInt(rng, minRoomSize, maxRoomSize)

		// Calculate position based on direction.
		// NOTE: Right now we center the neighboring room.
		//
		// This could be changed to use a random offset to
		// make the rooms more varied.
		var x, y int
		switch dir {
		case DirNorth:
			x = room.X + room.W/2 - rw/2
			y = room.Y - rl - 1
		case DirEast:
			x = room.X + room.W + 1
			y = room.Y + room.H/2 - rl/2
		case DirSouth:
			x = room.X + room.W/2 - rw/2
			y = room.Y + room.H + 1
		case DirWest:
			x = room.X - rw - 1
			y = room.Y + room.H/2 - rl/2
		}

		// Create a new room.
		newRoom := &Room{
			X: x,
			Y: y,
			W: rw,
			H: rl,
		}

		// Check if the new room overlaps with any existing rooms.
		if newRoom.X < 1 || newRoom.Y < 1 || newRoom.X+newRoom.W > width-1 || newRoom.Y+newRoom.H > height-1 || newRoom.Overlaps(rooms) {
			continue
		}

		// Append the room to the dungeon.
		rooms = append(rooms, newRoom)

		// Draw room.
		w.CarveRoom(newRoom)

		// There is a chance that a creature entity is placed randomly
		// in the room.

		// Pick a random location in the room.
		cx := randInt(rng, newRoom.X, newRoom.X+newRoom.W)
		cy := randInt(rng, newRoom.Y, newRoom.Y+newRoom.H)
		w.Entities = append(w.Entities, NewEntity(cx, cy, MonsterEntities[rng.Intn(len(MonsterEntities))]))

		// Draw a tunnel between the rooms.
		// NOTE: Right now, we just place the door in the middle.
		switch dir {
		case DirNorth:
			w.Cells[room.Y-1][room.X+room.W/2] = ' '
		case DirEast:
			w.Cells[room.Y+room.H/2][room.X+room.W] = ' '
		case DirSouth:
			w.Cells[room.Y+room.H][room.X+room.W/2] = ' '
		case DirWest:
			w.Cells[room.Y+room.H/2][room.X-1] = ' '
		}
		// Stop if we have enough rooms.
		if len(rooms) > maxRooms {
			break
		}
	}
	return w
}

// Room represents a room in the world.
// TODO: Store connecting rooms
type Room struct {
	X, Y int // top left corner
	W, H int // width and height
}

// Overlaps returns true if the given room overlaps with any of the rooms in the list.
func (r *Room) Overlaps(rooms []*Room) bool {
	for _, room := range rooms {
		if r.X+r.W < room.X || r.X > room.X+room.W {
			continue
		}
		if r.Y+r.H < room.Y || r.Y > room.Y+room.H {
			continue
		}
		return true
	}
	return false
}

// randInt returns a random integer between min and max using the given rng.
func randInt(rng *rand.Rand, min, max int) int {
	return min + rng.Intn(max-min)
}

// GenWorldBigBox generates a big box world.
func GenWorldBigBox(width, height int, seed int64) *World {
	w := NewWorld(width, height)
	w.Fill('#')
	w.CarveRoom(&Room{
		X: 1,
		Y: 1,
		W: width - 2,
		H: height - 2,
	})
	return w
}
