package genworldvoronoi

import (
	"image/color"
	//"log"
	"math"
)

// getTempFalloffFromAltitude returns the temperature falloff at a given altitude in meters
// above sea level. (approx. 9.8 °C per 1000 m)
// NOTE: This is definitely not correct :)
// Source: https://www.quora.com/At-what-rate-does-temperature-drop-with-altitude
func getTempFalloffFromAltitude(height float64) float64 {
	return 0.98 * height / 100
}

// getMeanAnnualTemp returns the temperature at a given latitude within the range of
// -35 °C to +31 °C.
// For this I assume that light hits the globe exactly from a 90° angle with respect
// to the planitary axis.
// See: https://www.scratchapixel.com/lessons/3d-basic-rendering/introduction-to-shading/shading-normals (facing ratio)
func getMeanAnnualTemp(lat float64) float64 {
	return (math.Sin(degToRad(90-math.Abs(lat))))*66 - 35
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}

// Biomes definition
// See: http://www-cs-students.stanford.edu/~amitp/game-programming/polygon-map-generation/#biomes
//
//
// Elevation ||| Moisture Zone ->
// Zone      V|| 6 (wet)    |    5    |     4    |     3    |     2     |   1 (dry)  |
// ==================================================================================
// 4 (high)   || SNOW ---------------------------| TUNDRA --| BARE -----|  SCORCHED -|
// ----------------------------------------------------------------------------------
// 3          || TAIGA ---------------| SHRUBLAND ----------| TEMPERATE DESERT ------|
// ----------------------------------------------------------------------------------
// 2          || TEMPERATE -| TEMPERATE ---------| GRASSLAND -----------| TEMPERATE -|
//            || RAIN FOREST| SEASONAL FOREST ---| ---------------------| DESERT ----|
// ----------------------------------------------------------------------------------
// 1 (low)    || TROPICAL RAIN FOREST | TROPICAL SEASONAL --| GRASSLAND | SUBTROPICAL|
//            || ---------------------| FOREST -------------| ----------| DESERT ----|
// ----------------------------------------------------------------------------------

var RedblobLookupTable = [][]int{
	{0x7, 0xA, 0x8, 0x8, 0x5, 0x5},
	{0x9, 0xA, 0xA, 0x8, 0x8, 0x5},
	{0x9, 0x9, 0x3, 0x3, 0x2, 0x2},
	{0xD, 0xC, 0x1, 0xB, 0xB, 0xB},
}

const (
	RedblobBiomeUnknown                 = 0x0
	RedblobBiomeTundra                  = 0x1
	RedblobBiomeTaiga                   = 0x2
	RedblobBiomeShrubland               = 0x3
	RedblobBiomeTemperateRainForest     = 0x4
	RedblobBiomeTropicalRainForest      = 0x5
	RedblobBiomeTropicalSeasonalForest  = 0x6
	RedblobBiomeSubtropicalDesert       = 0x7
	RedblobBiomeTemperateSeasonalForest = 0x8
	RedblobBiomeTemperateDesert         = 0x9 // combined with BioGrassland in whittaker
	RedblobBiomeGrassland               = 0xA // combined with BioTemperateDesert in whittaker
	RedblobBiomeSnow                    = 0xB // not in whittaker
	RedblobBiomeBare                    = 0xC // not in whittaker
	RedblobBiomeScorched                = 0xD // not in whittaker
)

func GetRedblobBiome(height, moisture int) int {
	height = RedblobClampHeightToIndex(height)
	moisture = RedblobClampMoistureToIndex(moisture)
	return RedblobLookupTable[height][moisture]
}

var RedblobBiomeColor = map[int]color.NRGBA{
	RedblobBiomeSnow:                    {0xFF, 0xFF, 0xFF, 0},
	RedblobBiomeTundra:                  {0xDD, 0xDD, 0xBB, 0},
	RedblobBiomeBare:                    {0xBB, 0xBB, 0xBB, 0},
	RedblobBiomeScorched:                {0x99, 0x99, 0x99, 0},
	RedblobBiomeTaiga:                   {0xCC, 0xD4, 0xBB, 0},
	RedblobBiomeShrubland:               {0xC4, 0xCC, 0xBB, 0},
	RedblobBiomeTemperateDesert:         {0xE4, 0xE8, 0xCA, 0},
	RedblobBiomeTemperateRainForest:     {0xA4, 0xC4, 0xA8, 0},
	RedblobBiomeTemperateSeasonalForest: {0xB4, 0xC9, 0xA9, 0},
	RedblobBiomeTropicalRainForest:      {0x9C, 0xBB, 0xA9, 0},
	RedblobBiomeTropicalSeasonalForest:  {0xA9, 0xCC, 0xA4, 0},
	RedblobBiomeGrassland:               {0xC4, 0xD4, 0xAA, 0},
	RedblobBiomeSubtropicalDesert:       {0xE9, 0xDD, 0xC7, 0},
}

func GetRedblobBiomeColor(height, moisture int, intensity float64) color.NRGBA {
	c := RedblobBiomeColor[GetRedblobBiome(height, moisture)]
	return color.NRGBA{
		R: uint8(intensity * float64(c.R)),
		G: uint8(intensity * float64(c.G)),
		B: uint8(intensity * float64(c.B)),
		A: 255,
	}
}

func RedblobClampHeightToIndex(height int) int {
	if height <= 4 && height >= 1 {
		return height - 1
	}
	if height < 1 {
		return 0
	}
	return 3
}

func RedblobClampMoistureToIndex(moisture int) int {
	if moisture <= 6 && moisture >= 1 {
		return moisture - 1
	}
	if moisture < 1 {
		return 0
	}
	return 5
}

// Whittaker biomes based on:
// https://github.com/JoeyR/FastBiome
var WhittakerLookupTable = [][]int{
	// <- +30°C                                                                                                                                         | 0 °C                                                                     |<- -15°C
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9}, // 0dm precipitation
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x8, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0x9, 0x0, 0x0, 0x0, 0x0},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0x9, 0x0, 0x0, 0x0, 0x0},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x2, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x2, 0x2, 0x2, 0x2, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x3, 0x3, 0x3, 0x3, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x3, 0x3, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x8, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x8, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x8, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x8, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x2, 0x2, 0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x0, 0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x0, 0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x0, 0x0, 0x0, 0x1, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	{0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, // 45dm precipitation
}

const (
	WhittakerBiomeUnknown                       = 0x0
	WhittakerBiomeTropicalRainForest            = 0x1
	WhittakerBiomeTropicalSeasonalForestSavanna = 0x2
	WhittakerBiomeSubtropicalDesert             = 0x3
	WhittakerBiomeTemperateRainForest           = 0x4
	WhittakerBiomeTemperateSeasonalForest       = 0x5
	WhittakerBiomeWoodlandShrubland             = 0x6
	WhittakerBiomeTemperateGrasslandDesert      = 0x7
	WhittakerBiomeBorealForestTaiga             = 0x8
	WhittakerBiomeTundra                        = 0x9
)

func GetWhittakerBiome(temperature, precipitation int) int {
	temperature = WhittakerClampTemperatureToIndex(temperature)
	precipitation = WhittakerClampPrecipitationToIndex(precipitation)
	return WhittakerLookupTable[precipitation][temperature]
}

var WhittakerBiomeColor = map[int]color.NRGBA{
	WhittakerBiomeTropicalRainForest:            {0x9C, 0xBB, 0xA9, 0},
	WhittakerBiomeTropicalSeasonalForestSavanna: {0xA9, 0xCC, 0xA4, 0},
	WhittakerBiomeSubtropicalDesert:             {0xE9, 0xDD, 0xC7, 0},
	WhittakerBiomeTemperateRainForest:           {0xA4, 0xC4, 0xA8, 0},
	WhittakerBiomeTemperateSeasonalForest:       {0xB4, 0xC9, 0xA9, 0},
	WhittakerBiomeWoodlandShrubland:             {0xC4, 0xCC, 0xBB, 0},
	WhittakerBiomeTemperateGrasslandDesert:      {0xE4, 0xE8, 0xCA, 0},
	WhittakerBiomeBorealForestTaiga:             {0xCC, 0xD4, 0xBB, 0},
	WhittakerBiomeTundra:                        {0xDD, 0xDD, 0xBB, 0},
}

func GetWhittakerBiomeColor(temperature, precipitation int, intensity float64) color.NRGBA {
	c := WhittakerBiomeColor[GetWhittakerBiome(temperature, precipitation)]
	return color.NRGBA{
		R: uint8(intensity * float64(c.R)),
		G: uint8(intensity * float64(c.G)),
		B: uint8(intensity * float64(c.B)),
		A: 255,
	}
}

var WhittakerLookupTableModified = [][]int{
	// <- +30°C                                                                                                                                         | 0 °C                                                                     |<- -15°C
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0xA, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9}, // 0dm precipitation
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x8, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0x9, 0xB, 0xB, 0xB, 0xB},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x7, 0x7, 0x7, 0x7, 0x7, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0x9, 0xB, 0xB, 0xB, 0xB},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0xE, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0xE, 0xE, 0xE, 0xE, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0x9, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0x9, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x3, 0x3, 0x3, 0x3, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0x9, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x3, 0x3, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x9, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0x2, 0x2, 0x2, 0x2, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xE, 0xE, 0xE, 0xE, 0xE, 0xE, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x6, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xE, 0xE, 0xE, 0xE, 0xE, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xE, 0xE, 0xE, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xE, 0xE, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x8, 0x8, 0x8, 0x8, 0x8, 0x8, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xE, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x8, 0x8, 0x8, 0x8, 0x8, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x8, 0x8, 0x8, 0x8, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x8, 0x8, 0x8, 0x8, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x8, 0x8, 0x8, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x8, 0x8, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x5, 0x5, 0x5, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x2, 0x2, 0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0x4, 0x4, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x4, 0xC, 0xC, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xD, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xD, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xD, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xD, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xC, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xC, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xC, 0xD, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xC, 0xC, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xC, 0xC, 0x1, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xC, 0xC, 0xC, 0x1, 0x1, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xD, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
	{0xC, 0xC, 0xC, 0xC, 0x1, 0x1, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0xC, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xD, 0xD, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB}, // 45dm precipitation
}

const (
	WhittakerModBiomeUnknown                 = 0x0
	WhittakerModBiomeTropicalRainForest      = 0x1
	WhittakerModBiomeTropicalSeasonalForest  = 0x2
	WhittakerModBiomeSubtropicalDesert       = 0x3
	WhittakerModBiomeTemperateRainForest     = 0x4
	WhittakerModBiomeTemperateSeasonalForest = 0x5
	WhittakerModBiomeWoodlandShrubland       = 0x6
	WhittakerModBiomeTemperateGrassland      = 0x7
	WhittakerModBiomeBorealForestTaiga       = 0x8
	WhittakerModBiomeTundra                  = 0x9
	WhittakerModBiomeColdDesert              = 0xA
	WhittakerModBiomeSnow                    = 0xB
	WhittakerModBiomeHotSwamp                = 0xC
	WhittakerModBiomeWetlands                = 0xD
	WhittakerModBiomeSavannah                = 0xE
)

func GetWhittakerModBiome(temperature, precipitation int) int {
	temperature = WhittakerClampTemperatureToIndex(temperature)
	precipitation = WhittakerClampPrecipitationToIndex(precipitation)
	return WhittakerLookupTableModified[precipitation][temperature]
}

var WhittakerModBiomeColor = map[int]color.NRGBA{
	WhittakerModBiomeTropicalRainForest:      {0x9C, 0xBB, 0xA9, 0},
	WhittakerModBiomeTropicalSeasonalForest:  {0xA9, 0xCC, 0xA4, 0},
	WhittakerModBiomeSubtropicalDesert:       {0xfd, 0xe3, 0x8d, 0}, // was 0xE9DDC7
	WhittakerModBiomeTemperateRainForest:     {0xA4, 0xC4, 0xA8, 0},
	WhittakerModBiomeTemperateSeasonalForest: {0xB4, 0xC9, 0xA9, 0},
	WhittakerModBiomeWoodlandShrubland:       {0xC4, 0xCC, 0xBB, 0},
	WhittakerModBiomeTemperateGrassland:      {0xE4, 0xE8, 0xCA, 0},
	WhittakerModBiomeBorealForestTaiga:       {0xCC, 0xD4, 0xBB, 0},
	WhittakerModBiomeTundra:                  {0xDD, 0xDD, 0xBB, 0},
	WhittakerModBiomeColdDesert:              {0xCC, 0xCC, 0xCC, 0},
	WhittakerModBiomeHotSwamp:                {0x96, 0x4B, 0x00, 0},
	WhittakerModBiomeSavannah:                {0xff, 0xd1, 0x45, 0},
	WhittakerModBiomeSnow:                    {0xFF, 0xFF, 0xFF, 0},
}

func GetWhittakerModBiomeColor(temperature, precipitation int, intensity float64) color.NRGBA {
	c := WhittakerModBiomeColor[GetWhittakerModBiome(temperature, precipitation)]
	return color.NRGBA{
		R: uint8(intensity * float64(c.R)),
		G: uint8(intensity * float64(c.G)),
		B: uint8(intensity * float64(c.B)),
		A: 255,
	}
}

const (
	minTemperatureC    = -15
	maxTemperatureC    = 30
	minPrecipitationDM = 0
	maxPrecipitationDM = 45
)

func WhittakerClampTemperatureToIndex(temp int) int {
	if temp <= minTemperatureC {
		return 44
	}
	if temp >= minTemperatureC {
		return 0
	}
	return 44 - (temp + 15)
}

func WhittakerClampPrecipitationToIndex(prec int) int {
	if prec >= maxPrecipitationDM {
		return 44
	}
	if prec <= minPrecipitationDM {
		return 0
	}
	return prec
}
