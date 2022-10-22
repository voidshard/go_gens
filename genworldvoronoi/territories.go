package genworldvoronoi

import (
	"container/heap"
	"container/list"
	"log"
	"math"
)

// Empire contains information about a territory with the given ID.
// TODO: Maybe drop the regions since we can get that info
// relatively cheaply.
type Empire struct {
	ID       int // ID of the territory
	Name     string
	Capital  *City   // Capital city
	Cities   []*City // Cities within the territory
	Regions  []int   // Regions that are part of the empire
	Language *Language
	ResMetal [ResMaxMetals]int
	ResGems  [ResMaxGems]int
}

func (e *Empire) Log() {
	log.Printf("The Empire of %s: %d cities, %d regions, capital: %s", e.Name, len(e.Cities), len(e.Regions), e.Capital.Name)
	for i := 0; i < ResMaxMetals; i++ {
		log.Printf("Metal %s: %d", metalToString(i), e.ResMetal[i])
	}
	for i := 0; i < ResMaxGems; i++ {
		log.Printf("Gem %s: %d", gemToString(i), e.ResGems[i])
	}
}

func (m *Map) GetEmpires() []*Empire {
	var res []*Empire
	for i := 0; i < m.NumTerritories; i++ {
		lang := GenLanguage(m.seed + int64(i))
		e := &Empire{
			ID:       m.cities_r[i].R,
			Name:     lang.MakeName(),
			Capital:  m.cities_r[i],
			Language: lang,
		}

		// TODO: Name empire, name cities.

		// Loop through all cities and gather all that
		// are within the current territory.
		for _, c := range m.cities_r {
			if m.r_territory[c.R] == e.ID {
				c.Name = e.Language.MakeName()
				e.Cities = append(e.Cities, c)
			}
		}

		// Collect all regions that are part of the
		// current territory.
		for r, terr := range m.r_territory {
			if terr == e.ID {
				for i := 0; i < ResMaxMetals; i++ {
					if m.res_metals_r[r]&(1<<i) != 0 {
						e.ResMetal[i]++
					}
				}
				for i := 0; i < ResMaxGems; i++ {
					if m.res_gems_r[r]&(1<<i) != 0 {
						e.ResGems[i]++
					}
				}
				e.Regions = append(e.Regions, r)
			}
		}
		e.Log()
		res = append(res, e)
	}
	return res
}

// identifyLandmasses returns a mapping from region to landmass ID.
// A landmass is a connected number of regions above sealevel.
func (m *Map) identifyLandmasses() []int {
	// NOTE: this is still in need of refinement.
	landMasses := make([]int, m.mesh.numRegions)
	for r := range landMasses {
		landMasses[r] = -1
	}
	var landID int
	var landSizes []int
	for r, h := range m.r_elevation {
		// Skip if the current region has already been allocated
		// or is below sealevel.
		if landMasses[r] != -1 || h < 0 {
			continue
		}
		var currentLandSize int
		queue := list.New()
		enqueue := func(r int) {
			// Skip if the current region has already been allocated
			// or is below sealevel.
			if landMasses[r] != -1 || m.r_elevation[r] < 0 {
				return
			}
			landMasses[r] = landID // Assign current landID to the region.
			currentLandSize++      // Increase size of known landmass.
			for _, nb := range m.rNeighbors(r) {
				// Skip if the neighbor region has already been allocated
				// or is below sealevel.
				if landMasses[nb] != -1 || m.r_elevation[nb] < 0 {
					continue
				}
				queue.PushBack(nb)
			}
		}
		// Start queue with current region.
		queue.PushBack(r)

		// Process each queue entry until we run out of
		// regions that belong to this landmass.
		for queue.Len() > 0 {
			e := queue.Front()
			enqueue(e.Value.(int))
			queue.Remove(e)
		}

		// Once done, append the current size to the list of landmass-
		// sizes and increment the current landID.
		landSizes = append(landSizes, currentLandSize)
		landID++
	}
	log.Println("number of landmasses", landID)
	log.Println(landSizes)
	return landMasses
}

type queueRegionEntry struct {
	index int // The index of the item in the heap.
	score float64
	city  int
	vx    int
}

// territoryQueue implements heap.Interface and holds Items.
type territoryQueue []*queueRegionEntry

func (pq territoryQueue) Len() int { return len(pq) }

func (pq territoryQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	// return pq[i].score > pq[j].score // 3, 2, 1
	return pq[i].score < pq[j].score // 1, 2, 3
}

func (pq territoryQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}

func (pq *territoryQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*queueRegionEntry)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *territoryQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (m *Map) rPlaceNTerritories(n int) {
	// Get maxFlux and maxElev for normalizing.
	_, maxFlux := minMax(m.r_flux)
	_, maxElev := minMax(m.r_elevation)

	// Truncate the number of territories to the number
	// of cities, just in case we have less cities than
	// territories.
	if n > len(m.cities_r) {
		n = len(m.cities_r)
	}
	var queue territoryQueue
	heap.Init(&queue)
	weight := func(u, v int) float64 {
		// Don't cross from water to land and vice versa.
		if (m.r_elevation[u] > 0) != (m.r_elevation[v] > 0) {
			return -1
		}

		// Calculate horizontal distance.
		ulat := m.r_latLon[u][0]
		ulon := m.r_latLon[u][1]
		vlat := m.r_latLon[v][0]
		vlon := m.r_latLon[v][1]
		horiz := haversine(ulat, ulon, vlat, vlon) / (2 * math.Pi)

		// Calculate vertical distance.
		vert := (m.r_elevation[v] - m.r_elevation[u]) / maxElev
		if vert > 0 {
			vert /= 10
		}
		diff := 1 + 0.25*math.Pow(vert/horiz, 2)
		diff += 100 * math.Sqrt(m.r_flux[u]/maxFlux)
		if m.r_elevation[u] <= 0 {
			diff = 100
		}
		return horiz * diff
	}

	// 'terr' will hold a mapping of region to territory.
	// The territory ID is the region number of the capital city.
	terr := make([]int, m.mesh.numRegions)
	for i := 0; i < n; i++ {
		terr[m.cities_r[i].R] = m.cities_r[i].R
		for _, v := range m.rNeighbors(m.cities_r[i].R) {
			if m.r_elevation[v] <= 0 {
				continue
			}
			heap.Push(&queue, &queueRegionEntry{
				score: weight(m.cities_r[i].R, v),
				city:  m.cities_r[i].R,
				vx:    v,
			})
		}
	}

	// Extend territories until the queue is empty.
	for queue.Len() > 0 {
		u := heap.Pop(&queue).(*queueRegionEntry)
		if terr[u.vx] != 0 {
			continue
		}
		terr[u.vx] = u.city
		for _, v := range m.rNeighbors(u.vx) {
			if terr[v] != 0 || m.r_elevation[v] < 0 {
				continue
			}
			newdist := weight(u.vx, v)
			if newdist < 0 {
				continue
			}
			heap.Push(&queue, &queueRegionEntry{
				score: u.score + newdist,
				city:  u.city,
				vx:    v,
			})
		}
	}
	m.r_territory = terr
}
