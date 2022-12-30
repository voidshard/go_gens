package genworldvoronoi

import (
	"log"
	"math"
)

// sumResources returns the sum of the resource flag IDs in the byte.
// This is a convenience function for determining the approximate
// value of local resources.
// NOTE: In theory one could just cast the int to a byte and use the
// value like that, but the value would be a power of 2, which might
// be too stark a difference.
func sumResources(r byte) int {
	sum := 0
	for i := 0; i < 8; i++ {
		if r&(1<<i) != 0 {
			sum += i + 1
		}
	}
	return sum
}

const (
	ResourceTypeMetal = iota
	ResourceTypeGem
	ResourceTypeStone
)

// getRegsWithResource returns the regions that have the specified resource.
func (m *Geo) getRegsWithResource(resource byte, resourceType int) []int {
	// Pick the correct resource slice.
	var search []byte
	switch resourceType {
	case ResourceTypeMetal:
		search = m.Metals
	case ResourceTypeGem:
		search = m.Gems
	case ResourceTypeStone:
		search = m.Stones
	}

	// Find the regions that have the specified resource.
	var regions []int
	for r, val := range search {
		if val&resource != 0 {
			regions = append(regions, r)
		}
	}
	return regions
}

// Resources maps regions to natural resources.
type Resources struct {
	Metals []byte // Metal ores
	Gems   []byte // Gemstones
	Stones []byte // Rocks or minerals
}

func newResources(size int) *Resources {
	return &Resources{
		Metals: make([]byte, size),
		Gems:   make([]byte, size),
		Stones: make([]byte, size),
	}
}

func (m *Geo) resourceFitness() []float64 {
	fitness := make([]float64, m.mesh.numRegions)
	f := m.getFitnessSteepMountains()
	for r := range fitness {
		fitness[r] = f(r)
	}
	return fitness
}

func (m *Geo) placeResources() {
	// NOTE: This currently sucks.
	// TODO: Use fitness function instead or in addition.

	// Place metals.
	// Metals can be found mainly in mountains, so steepness
	// will be an indicator along with the distance from the
	// mountain seed points.
	m.placeMetals()

	// Place gemstones.
	// Gemstones can be found mainly in inland valleys, so
	// distance from the coastlines, mountains, and oceans
	// will be an indicator.
	m.placeGems()

	// Place forests.
	// Forests can be found mainly in valleys, so steepness
	// will be an indicator along with the distance from the
	// valley's center.
	m.placeForests()

	// Place potential quarry sites.
	// Potential quarry sites can be found mainly in mountains,
	m.placeStones()

	// Place energy sources.
	// Oil, coal, and natural gas, as well as geothermal energy
	// and magical handwavium.

	// Place arable land.
	// Arable land can be found mainly in valleys, so steepness
	// will be an indicator along with the distance from the
	// valley's center.
}

// Metal resource flags starting with the cheapest metal.
const (
	ResMetIron = 1 << iota
	ResMetCopper
	ResMetLead
	ResMetTin
	ResMetSilver
	ResMetGold
	ResMetPlatinum
)

const ResMaxMetals = 7

func metalToString(metalID int) string {
	switch 1 << metalID {
	case ResMetIron:
		return "Iron"
	case ResMetCopper:
		return "Copper"
	case ResMetLead:
		return "Lead"
	case ResMetTin:
		return "Tin"
	case ResMetSilver:
		return "Silver"
	case ResMetGold:
		return "Gold"
	case ResMetPlatinum:
		return "Platinum"
	default:
		return "Unknown"
	}
}

func (m *Geo) placeMetals() {
	steepness := m.GetSteepness()
	// distMountains, _, _, _ := m.findCollisions()

	// https://www.reddit.com/r/worldbuilding/comments/kbmnd6/a_guide_to_placing_resources_on_fictional_worlds/
	const (
		chancePlatinum = 0.005
		chanceGold     = chancePlatinum + 0.020
		chanceSilver   = chanceGold + 0.040
		chanceCopper   = chanceSilver + 0.06
		chanceLead     = chanceCopper + 0.07
		chanceTin      = chanceLead + 0.1
		chanceIron     = chanceTin + 0.4
	)
	fn := m.fbmNoiseCustom(2, 1, 2, 2, 2, 0, 0, 0)
	fm := m.getFitnessSteepMountains()

	// NOTE: By encoding the resources as bit flags, we can easily
	// determine the value of a region given the assumption that
	// each resource is twice (or half) as valuable as the previous
	// resource. This will be handy for fitness functions and such.
	//
	// I feel pretty clever about this one, but it's not realistic.
	m.resetRand()
	metals := make([]byte, len(steepness))

	// TODO: Use noise intersection instead of rand.
	for r := 0; r < m.mesh.numRegions; r++ {
		if fm(r) > 0.9 {
			switch rv := math.Abs(m.rand.NormFloat64() * fn(r)); {
			case rv < chancePlatinum:
				metals[r] |= ResMetPlatinum
			case rv < chanceGold:
				metals[r] |= ResMetGold
			case rv < chanceSilver:
				metals[r] |= ResMetSilver
			case rv < chanceCopper:
				metals[r] |= ResMetCopper
			case rv < chanceLead:
				metals[r] |= ResMetLead
			case rv < chanceTin:
				metals[r] |= ResMetTin
			case rv < chanceIron:
				metals[r] |= ResMetIron
			}
		}
	}
	m.Metals = metals

	// This attempts some weird variation of:
	// https://www.redblobgames.com/x/1736-resource-placement/
	/*
		nA := m.fbm_noise2(5, 0.5, 5, 5, 5, 0, 0, 0)
		nB := m.fbm_noise2(7, 0.5, 5, 5, 5, 0, 0, 0)
		resources := make([]byte, len(steepness))
		for r := range steepness {
			noiseVal := (nA(r) + nB(r) + m.r_elevation[r]) / 3
			if m.getIntersection(noiseVal, 0.75, 0.01) {
				resources[r] |= ResMetPlatinum
			}
			//chance /= float64(distMountains[r])
		}

		nC := m.fbm_noise2(2, 0.5, 5, 5, 5, 0, 0, 0)
		nD := m.fbm_noise2(7, 0.5, 5, 5, 5, 0, 0, 0)
		for r := range steepness {
			noiseVal := (nC(r) + nD(r) + m.r_elevation[r]) / 3
			if m.getIntersection(noiseVal, 0.75, 0.02) {
				resources[r] |= ResMetGold
			}
			//chance /= float64(distMountains[r])
		}

		nC = m.fbm_noise2(2, 0.5, 1, 1, 1, 0, 0, 0)
		nD = m.fbm_noise2(5, 0.1, 1, 1, 1, 0, 0, 0)
		for r := range steepness {
			noiseVal := (-1*(nC(r)+nD(r)) + m.r_elevation[r]) / 3
			if m.getIntersection(noiseVal, 0.52, 0.07) {
				resources[r] |= ResMetIron
			}
			//chance /= float64(distMountains[r])
		}
	*/

	//m.r_metals = resources
}

// Gemstone resource flags starting with the cheapest gem.
const (
	ResGemAmethyst = 1 << iota
	ResGemTopaz
	ResGemSapphire
	ResGemEmerald
	ResGemRuby
	ResGemDiamond
)

const ResMaxGems = 6

func gemToString(gemsID int) string {
	switch 1 << gemsID {
	case ResGemAmethyst:
		return "Amethyst"
	case ResGemTopaz:
		return "Topaz"
	case ResGemSapphire:
		return "Sapphire"
	case ResGemEmerald:
		return "Emerald"
	case ResGemRuby:
		return "Ruby"
	case ResGemDiamond:
		return "Diamond"
	default:
		return "Unknown"
	}
}

func (m *Geo) placeGems() {
	steepness := m.GetSteepness()
	const (
		chanceDiamond  = 0.005
		chanceRuby     = chanceDiamond + 0.025
		chanceEmerald  = chanceRuby + 0.04
		chanceSapphire = chanceEmerald + 0.05
		chanceTopaz    = chanceSapphire + 0.06
		chanceAmethyst = chanceTopaz + 0.1
		// chanceQuartz   = 0.75 // Usually goes hand in hand with gold?
		// chanceFlint    = 0.9
	)

	gems := make([]byte, len(steepness))
	for r := 0; r < m.mesh.numRegions; r++ {
		if steepness[r] > 0.9 && m.Elevation[r] > 0.5 {
			switch rv := math.Abs(m.rand.NormFloat64()); {
			case rv < chanceDiamond:
				gems[r] |= ResGemDiamond
			case rv < chanceRuby:
				gems[r] |= ResGemRuby
			case rv < chanceEmerald:
				gems[r] |= ResGemEmerald
			case rv < chanceSapphire:
				gems[r] |= ResGemSapphire
			case rv < chanceTopaz:
				gems[r] |= ResGemTopaz
			case rv < chanceAmethyst:
				gems[r] |= ResGemAmethyst
				// case rv < chanceQuartz:
				//	gems[r] |= ResGemQuartz
				// case rv < chanceFlint:
				//	gems[r] |= ResGemFlint
			}
		}
	}
	m.Gems = gems
}

// Stone resource flags starting with the most common stone.
// NOTE: Clay?
const (
	ResStoSandstone = 1 << iota
	ResStoLimestone
	ResStoChalk
	ResStoMarble
	ResStoGranite
	ResStoBasalt
	ResStoObsidian
)

const ResMaxStones = 7

func stoneToString(stoneID int) string {
	switch 1 << stoneID {
	case ResStoSandstone:
		return "Sandstone"
	case ResStoLimestone:
		return "Limestone"
	case ResStoChalk:
		return "Chalk"
	case ResStoMarble:
		return "Marble"
	case ResStoGranite:
		return "Granite"
	case ResStoBasalt:
		return "Basalt"
	case ResStoObsidian:
		return "Obsidian"
	default:
		return "Unknown"
	}
}

const (
	ResVarClay = 1 << iota
	ResVarSulfur
	ResVarSalt
	ResVarCoal
	ResVarOil
	ResVarGas
)

func (m *Geo) placeStones() {
	log.Println("placing stones is not implemented")
}

func (m *Geo) placeForests() {
	log.Println("placing forests is not implemented")
	// Get all biomes that are forested.
	// Place trees in those biomes based on the biome's tree type(s).
	// Profit!
}
