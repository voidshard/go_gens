package genmap2derosion

import (
	"log"
	"math/rand"
	"time"

	opensimplex "github.com/ojrac/opensimplex-go"
)

func (w *World) genClimate() *Climate {
	dimX := w.params.Size.X
	dimY := w.params.Size.Y

	// Initialize terrain heightmap.
	//
	// Generate new heightmap from actual heightmap, scaled to
	// the expected height values.
	heightmap := make([]float64, len(w.heightmap))
	for i := range heightmap {
		heightmap[i] = w.heightmap[i]*4000 - 300
	}

	// Initialize climate.
	climate := NewClimate(int(dimX), int(dimY), 0, int(w.params.Seed), heightmap)

	// Calculate the climate system.
	climate.calcAverage()

	// Generate the surface composition.
	climate.genBiome()

	// Generate climate. This is really suboptimal.
	now := time.Now()

	if w.params.StorePNGCycles {
		w.ExportPng("b_image_terrain.png", heightmap)
	}

	// Run the simulation for 365 days.
	for day := 0; day < 365; day++ {
		// Run the simulation.
		log.Println(day)
		climate.runSimulation(day)

		if w.params.StoreGIFFrames {
			// Build a hacky float map that is supposed to represent
			// rain and clouds that we can export as a GIF frame.
			// TODO: Remove or improve.
			rm := make([]float64, len(climate.RainMap))
			for i, v := range climate.CloudMap {
				if v {
					rm[i] = 0.2
				} else {
					rm[i] = w.heightmap[i]
				}
			}
			for i, v := range climate.RainMap {
				if v {
					rm[i] = 0.7
				}
			}
			rm[0] = 0
			rm[1] = 1
			w.storeGifFrame(rm, heightmap, heightmap)
		}
	}

	if w.params.StorePNGCycles {
		w.ExportPng("b_image_avterrain.png", heightmap)
		w.ExportPng("b_image_avgrain.png", climate.AvgRainMap)
		w.ExportPng("b_image_avgtemp.png", climate.AvgTempMap)
		w.ExportPng("b_image_avgwind.png", climate.AvgWindMap)
		w.ExportPng("b_image_avgcloud.png", climate.AvgCloudMap)
	}

	log.Println(time.Since(now))
	return climate
}

type Climate struct {
	perlin     opensimplex.Noise // Open simplex which we pretend to be perlin
	seed       int               // Seed for Perlin Noise
	dimX, dimY int               // Dimensions of the map

	// Curent climate state.
	TempMap       []float64  // local temperature
	HumidityMap   []float64  // local humidity
	CloudMap      []bool     // cloud cover (true = cloudy)
	RainMap       []bool     // rain is falling (true = raining)
	WindMap       []float64  // local wind speeds
	WindDirection [2]float64 // global wind vector (from 0-1)

	// Average climate maps (collected over the course of the simulation)
	AvgRainMap     []float64 // average raininess over time
	AvgWindMap     []float64 // average wind speed over time
	AvgCloudMap    []float64 // average cloud cover over time
	AvgTempMap     []float64 // average temperature over time
	AvgHumidityMap []float64 // average humidity over time

	// Biome mapping.
	biomeMap []int

	// Heightmap.
	heightmap []float64
}

func NewClimate(dimX, dimY, day, seed int, heightmap []float64) *Climate {
	idxSize := dimX * dimY
	c := &Climate{
		dimX:           dimX,
		dimY:           dimY,
		seed:           seed,
		TempMap:        make([]float64, idxSize),
		HumidityMap:    make([]float64, idxSize),
		CloudMap:       make([]bool, idxSize),
		RainMap:        make([]bool, idxSize),
		WindMap:        make([]float64, idxSize),
		AvgRainMap:     make([]float64, idxSize),
		AvgWindMap:     make([]float64, idxSize),
		AvgCloudMap:    make([]float64, idxSize),
		AvgTempMap:     make([]float64, idxSize),
		AvgHumidityMap: make([]float64, idxSize),
		biomeMap:       make([]int, idxSize),
		heightmap:      heightmap,
	}
	c.init(day)
	return c
}

func (c *Climate) init(day int) {
	if c.perlin == nil {
		c.perlin = opensimplex.New(int64(c.seed))
	}

	// Initialize and calculate winds for the given day.
	c.calcWind(day)

	// Initialize temperature map.
	for i := range c.TempMap {
		// Add for height.
		if h := c.heightmap[i]; h > 200 {
			// NOTE: I'd much prefer to calculate the temperature using a proper
			// temperature falloff formula.
			c.TempMap[i] = 1 - h/2000
		} else {
			// Sealevel temperature.
			// In Degrees Celsius (NOTE: Is that right? Seems awfully low.)
			c.TempMap[i] = 0.7
		}
	}

	// Initialize humidity map.
	for i := range c.HumidityMap {
		// Humidty increases above water bodies (regions below sea level).
		if c.heightmap[i] < 200 {
			c.HumidityMap[i] = 0.4
		} else {
			c.HumidityMap[i] = 0.2
		}
	}

	// Initialize rain map.
	for i := range c.RainMap {
		c.RainMap[i] = false
	}

	// Initialize cloud map.
	for i := range c.CloudMap {
		c.CloudMap[i] = false
	}
}

func (c *Climate) runSimulation(day int) {
	// Calculate new wind speeds.
	c.calcWind(day)

	// Calculate the temperature map.
	c.calcTempMap()

	// Calculate the humidity map.
	c.calcHumidityMap()

	// Calculate the cloud map.
	c.calcRainMap()
}

// calcAverage calculates the average of the climate over a number of years.
func (c *Climate) calcAverage() {
	// Climate simulation over n years.
	years := 1
	startDay := 0

	// Initiate climate maps for averaging.
	for i := range c.heightmap {
		c.AvgRainMap[i] = 0
		c.AvgWindMap[i] = 0
		c.AvgCloudMap[i] = 0
		c.AvgTempMap[i] = 0
		c.AvgHumidityMap[i] = 0
	}

	// Initiate simulation at a starting point.
	simulation := NewClimate(c.dimX, c.dimY, startDay, c.seed, c.heightmap)

	// Simulate every day for a number of years.
	for i, days := 0, years*365; i < days; i++ {
		// Calculate new climate state.
		simulation.calcWind(i)
		simulation.calcTempMap()
		simulation.calcHumidityMap()
		simulation.calcRainMap()

		// Calculate moving average.
		for idx := range c.heightmap {
			// Average wind.
			c.AvgWindMap[idx] = calcMovingAverage(c.AvgWindMap[idx], simulation.WindMap[idx], i)

			// Average rain.
			c.AvgRainMap[idx] = calcMovingAverageBool(c.AvgRainMap[idx], simulation.RainMap[idx], i)

			// Average cloud cover.
			c.AvgCloudMap[idx] = calcMovingAverageBool(c.AvgCloudMap[idx], simulation.CloudMap[idx], i)

			// Average temperature.
			c.AvgTempMap[idx] = calcMovingAverage(c.AvgTempMap[idx], simulation.TempMap[idx], i)

			// Average humidity.
			c.AvgHumidityMap[idx] = calcMovingAverage(c.AvgHumidityMap[idx], simulation.HumidityMap[idx], i)

		}
	}
}

func calcMovingAverageBool(v float64, newv bool, i int) float64 {
	if newv {
		return calcMovingAverage(v, 1, i)
	}
	return calcMovingAverage(v, 0, i)
}

func calcMovingAverage(v, newv float64, i int) float64 {
	return (v*float64(i) + newv) / float64(i+1)
}

// calcWind sets the wind direction and calculates local wind speed for the given day.
func (c *Climate) calcWind(day int) {
	timeInterval := float64(day) / 365

	// Winddirection shifts every day (one dimensional noise).
	c.WindDirection[0] = (c.perlin.Eval2(timeInterval, float64(c.seed)))
	c.WindDirection[1] = (c.perlin.Eval2(timeInterval, timeInterval+float64(c.seed)))

	dx := c.dimX
	dy := c.dimY
	wdx := c.WindDirection[0]
	wdy := c.WindDirection[1]
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			idx := i*dy + j

			// Get previous tile index (upwind).
			// Assumption: Wind blows despite obstacles.
			k := i + int(10*wdx)
			if k < 0 || k >= dx {
				k = i
			}
			l := j + int(10*wdy)
			if l < 0 || l >= dy {
				l = j
			}

			c.WindMap[idx] = 5 * (1 - (c.heightmap[idx]-c.heightmap[k*dy+l])/1000)
		}
	}
}

func (c *Climate) calcHumidityMap() {
	oldHumidMap := make([]float64, c.dimX*c.dimY)
	// Copy humidity to old humidity map.
	copy(oldHumidMap, c.HumidityMap)

	dx := c.dimX
	dy := c.dimY
	wdx := c.WindDirection[0]
	wdy := c.WindDirection[1]
	for i := 1; i < dx-1; i++ {
		for j := 1; j < dy-1; j++ {
			idx := i*dy + j
			// Get new map index from wind direction

			// Get previous tile index (upwind).
			// Assumption: Wind blows despite obstacles.
			k := i + int(2*c.WindMap[idx]*wdx)
			if k < 0 || k >= dx {
				k = i
			}
			l := j + int(2*c.WindMap[idx]*wdy)
			if l < 0 || l >= c.dimY {
				l = j
			}

			// Transfer to New Tile
			c.HumidityMap[idx] = oldHumidMap[k*dy+l]

			// Average with neighbor values.
			newHumidity := (c.HumidityMap[idx-dy-1] +
				c.HumidityMap[idx+dy-1] +
				c.HumidityMap[idx+dy+1] +
				c.HumidityMap[idx-dy+1] +
				c.HumidityMap[idx] +
				c.HumidityMap[idx+1] +
				c.HumidityMap[idx-1] +
				c.HumidityMap[idx+dy] +
				c.HumidityMap[idx-dy]) / 9
			// newHumidity := (c.HumidityMap[idx-dy-1] + c.HumidityMap[idx+dy-1] + c.HumidityMap[idx+dy+1] + c.HumidityMap[idx-dy+1]) / 4

			// We are over a body of water, increased temperature due to
			// sunshine (unimpeded by cloud cover) adds humidity through
			// evaporation.
			var addHumidity float64
			if !c.CloudMap[idx] {
				if c.heightmap[idx] <= 200 {
					addHumidity = 0.05 * c.TempMap[idx]
				} else {
					addHumidity = 0.01
				}
			}

			// Raining
			var addRain float64
			if c.RainMap[idx] {
				addRain = -(newHumidity) * 0.8
			}

			newHumidity = newHumidity + (newHumidity)*addRain + (1-newHumidity)*(addHumidity)
			if newHumidity > 1 {
				newHumidity = 1
			} else if newHumidity < 0 {
				newHumidity = 0
			}
			c.HumidityMap[idx] = newHumidity
		}
	}
}

func (c *Climate) calcTempMap() {
	// Copy current temperatures to a temporary map.
	oldTempMap := make([]float64, c.dimX*c.dimY)
	copy(oldTempMap, c.TempMap)

	dx := c.dimX
	dy := c.dimY
	wdx := c.WindDirection[0]
	wdy := c.WindDirection[1]
	for i := 1; i < dx-1; i++ {
		for j := 1; j < dy-1; j++ {
			idx := i*dy + j

			// Get previous tile index (upwind).
			// Assumption: Wind blows despite obstacles.
			k := i + int(2*c.WindMap[idx]*wdx)
			if k < 0 || k >= dx {
				k = i
			}
			l := j + int(2*c.WindMap[idx]*wdy)
			if l < 0 || l >= dy {
				l = j
			}

			// Transfer from old to new tile.
			c.TempMap[idx] = oldTempMap[k*dy+l]

			// Average with neighbor values.
			newTemp := (c.TempMap[idx-dy-1] +
				c.TempMap[idx+dy-1] +
				c.TempMap[idx+dy+1] +
				c.TempMap[idx-dy+1] +
				c.TempMap[idx]) / 5

			// Various contributions to the TempMap

			// Rising air cools down.
			addCool := 0.5 * (c.WindMap[idx] - 5)

			// Sunlight on surface warms up.
			var addSun float64
			if !c.CloudMap[idx] {
				addSun = (1 - c.heightmap[idx]/2000) * 0.008
			}

			// Rain reduces temperature.
			var addRain float64
			if c.RainMap[idx] && newTemp > 0 {
				addRain = -0.01
			}

			// Add contributing factors.
			newTemp = newTemp + 0.8*(1-newTemp)*addSun + 0.6*(newTemp)*(addRain+addCool)

			// Clamp temperature between 0 and 1.
			if newTemp > 1 {
				newTemp = 1
			} else if newTemp < 0 {
				newTemp = 0
			}
			c.TempMap[idx] = newTemp
		}
	}
}

func (c *Climate) calcRainMap() {
	oldCloudMap := make([]bool, c.dimX*c.dimY)
	copy(oldCloudMap, c.CloudMap)

	oldRainMap := make([]bool, c.dimX*c.dimY)
	copy(oldRainMap, c.RainMap)

	for i := range oldRainMap {
		c.CloudMap[i] = false
		c.RainMap[i] = false
	}

	dx := c.dimX
	dy := c.dimY
	wdx := c.WindDirection[0]
	wdy := c.WindDirection[1]
	for i := 1; i < dx-1; i++ {
		for j := 1; j < dy-1; j++ {
			idx := i*dy + j

			// Get previous tile index (upwind).
			// Assumption: Wind blows despite obstacles.
			k := i + int(2*c.WindMap[idx]*wdx)
			if k < 0 || k >= dx {
				k = i
			}
			l := j + int(2*c.WindMap[idx]*wdy)
			if l < 0 || l >= dy {
				l = j
			}

			// Rain Condition.
			// Check if the humidity exceeds the seturation limit for the current
			// index. If so, it will start to rain.
			if c.HumidityMap[idx] >= 0.35+0.5*c.TempMap[idx] {
				c.RainMap[idx] = true
				c.CloudMap[idx] = oldCloudMap[k*dy+l] // Transfer to New Tile
			} else if c.HumidityMap[idx] >= 0.3+0.3*c.TempMap[idx] {
				c.CloudMap[idx] = true
				c.RainMap[idx] = oldRainMap[k*dy+l] // Transfer to New Tile
			} else {
				c.CloudMap[idx] = false
				c.RainMap[idx] = false
			}
		}
	}
}

func (c *Climate) genBiome() {
	// Determine the Surface Biome:
	// 0: Water
	// 1: Sandy Beach
	// 2: Gravel Beach
	// 3: Stone Beach Cliffs
	// 4: Wet Plains (Grassland)
	// 5: Dry Plains (Shrubland)
	// 6: Rocky Hills
	// 7: Tempererate Forest
	// 8: Boreal Forest
	// 9: Mountain Tundra
	// 10: Mountain Peak
	// Compare the Parameters and decide what kind of ground we have.
	for i := range c.heightmap {
		switch d := c.heightmap[i]; {
		case d <= 200:
			c.biomeMap[i] = 0 // 0: Water
		case d <= 204:
			c.biomeMap[i] = 1 // 1: Sandy Beach
		case d <= 210:
			c.biomeMap[i] = 2 // 2: Gravel Beach
		case d <= 220:
			c.biomeMap[i] = 3 // 3: Stony Beach Cliffs
		case d <= 600:
			if c.AvgRainMap[i] >= 0.02 {
				c.biomeMap[i] = 4 // 4: Wet Plains (Grassland)
			} else {
				c.biomeMap[i] = 5 // 5: Dry Plains (Shrubland)
			}
		case d <= 1300:
			x := i / c.dimY
			y := i % c.dimY
			if c.AvgRainMap[i] < 0.001 && x+rand.Int()%4-2 > 5 && x+rand.Int()%4-2 < 95 && y+rand.Int()%4-2 > 5 && y+rand.Int()%4-2 < 95 {
				c.biomeMap[i] = 6 //6: Rocky Hills
			} else if d <= 1100 {
				c.biomeMap[i] = 7 //7: Temperate Forest
			} else {
				c.biomeMap[i] = 8 //8: Boreal Forest
			}
		case d <= 1500:
			c.biomeMap[i] = 9
		default:
			c.biomeMap[i] = 10 //Otherwise just Temperate Forest
		}
	}
}

/*
type Vegetation struct {
}

func (v *Vegetation) getTree(territory World, player Player, i, j int) bool {
	//Code to Calculate wether or not we have a tree
	 Ideally this generates a vegetation map, spitting out
	0 for nothing,
	1 from short grass,
	2 for shrub,
	3 for some herb
	4 for some bush
	5 for some flower
	6 for some tree
	and also gives a number for a variant (3-5 variants of everything per biome)
	every variant could then also have a texture variant if wanted
	For now it only spits out wether or not we have a tree, which it then draws
	We can one piece of vegetation per map
	You could also do this for other objects on the map
	(tents, rocks, other locations) and not place vegetation if there is something present


	//Perlin Noise Module
	perlin := opensimplex.New(seed)
	// var perlin Perlin

	// perlin.SetOctaveCount(20)
	// perlin.SetFrequency(1000)
	// perlin.SetPersistence(0.8)

	//Generate the Height Map with Perlin Noise
	x := float64(player.xTotal-25+i) / 100000
	y := float64(player.yTotal-25+j) / 100000

	//This is not an efficient tree generation method
	//But a reasonable distribution for a grassland area
	//srand(x + y)
	tree := int((1/(perlin.Eval2(x, y, territory.seed+1)+1))*rand.Float64()%5) / 4

	return tree > 0
}*/
