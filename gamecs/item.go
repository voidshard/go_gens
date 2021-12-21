package gamecs

import (
	"github.com/Flokey82/go_gens/vectors"
)

type ItemLocation int

const (
	LocWorld ItemLocation = iota
	LocContainer
	LocInventory
)

type Item struct {
	Location ItemLocation
	Pos      vectors.Vec2
	*ItemType
	// Owned bool
	// TODO: Carryable? Maybe weight determines if one can carry it in one hand, two hands, an animal with its beak?
}

type ItemType struct {
	Name       string
	Tags       []string       // Food, Weapon
	Properties map[string]int // Price, weight, damage, ...
}

func newItemType(name string) *ItemType {
	return &ItemType{
		Name:       name,
		Properties: make(map[string]int),
	}
}

func (i *ItemType) New(pos vectors.Vec2) *Item {
	return &Item{
		ItemType: i,
		Pos:      pos,
	}
}
