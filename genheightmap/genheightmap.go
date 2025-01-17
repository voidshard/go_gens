package genheightmap

import (
	"math"
	"math/rand"

	"github.com/Flokey82/go_gens/vectors"

	opensimplex "github.com/ojrac/opensimplex-go"
)

// Terrain is an interface for a heightmap.
// ... I can't remember if I use this somewhere.
type Terrain interface {
	//ApplyGen(f GenFunc)
	MinMax() (float64, float64)
}

type GenFunc func(x, y float64) float64

// GenSlope returns a generator function that produces a slope in the direction
// of the given vector.
func GenSlope(direction vectors.Vec2) GenFunc {
	return func(x, y float64) float64 {
		return x*direction.X + y*direction.Y
	}
}

// GenCone returns a generator function for a cone at the center of the heightmap.
// TODO: Allow the user to specify the center of the cone.
func GenCone(slope float64) GenFunc {
	return func(x, y float64) float64 {
		return math.Pow(x*x+y*y, 0.5) * slope
	}
}

// GenVolCone returns a generator function for a volcanic cone
// at the center of the heightmap.
// TODO: Allow the user to specify the center of the cone.
func GenVolCone(slope float64) GenFunc {
	return func(x, y float64) float64 {
		dist := math.Pow(x*x+y*y, 0.5)
		if dist < 0.1 {
			return -4 * dist * slope
		}
		return dist * slope
	}
}

// GenMountains returns a generator function that will return the height of a
// point on the heightmap given the point's coordinates, which will produce a
// number of mountains.
// TODO: The seed should be passed into the function as parameter.
//
// 'maxX', 'maxY' are the dimensions of the heightmap.
// 'n' is the number of mountains.
// 'r' is the radius of the mountains.
func GenMountains(maxX, maxY float64, n int, r float64) GenFunc {
	rand.Seed(1234)
	var mounts [][2]float64
	for i := 0; i < n; i++ {
		mounts = append(mounts, [2]float64{maxX * (rand.Float64() - 0.5), maxY * (rand.Float64() - 0.5)})
	}
	return func(x, y float64) float64 {
		var val float64
		for j := 0; j < n; j++ {
			m := mounts[j]
			val += math.Pow(math.Exp(-((x-m[0])*(x-m[0])+(y-m[1])*(y-m[1]))/(2*r*r)), 2)
		}
		return val
	}
}

// GenNoise returns a function that returns the noise/height value of a given point
// on the heightmap. Not sure what the slope parameter was supposed to do.
func GenNoise(seed int64, slope float64) GenFunc {
	perlin := opensimplex.New(seed)

	mult := 15.0
	pow := 1.0
	return func(x, y float64) float64 {
		x *= mult
		y *= mult
		e := 1 * math.Abs(perlin.Eval2(x, y))
		e += 0.5 * math.Abs(perlin.Eval2(x*2, y*2))
		e += 0.25 * perlin.Eval2(x*4, y*4)
		e /= (1 + 0.5 + 0.25)
		return math.Pow(e, pow)
	}
}

// CalcMean calculates the mean of a slice of floats.
func CalcMean(nums []float64) float64 {
	total := 0.0
	for _, v := range nums {
		total += v
	}
	return total / float64(len(nums))
}

// MinMax returns the min and max values of the heightmap.
func MinMax(hm []float64) (float64, float64) {
	if len(hm) == 0 {
		return 0, 0
	}
	min := hm[0]
	max := hm[0]
	for _, h := range hm {
		if h > max {
			max = h
		}

		if h < min {
			min = h
		}
	}
	return min, max
}

// Modify is a function that modifies a value in a heightmap.
type Modify func(val float64) float64

// ModNormalize normalizes the heightmap to the range [0, 1] given
// the min and max values (the range of heightmap values).
func ModNormalize(min, max float64) Modify {
	return func(val float64) float64 {
		return (val - min) / (max - min)
	}
}

// ModPeaky returns the function applied to a point on a heightmap
// in order to exaggerate the peaks of the map.
func ModPeaky() Modify {
	return math.Sqrt
}

// ModSeaLevel shifts the origin point to the sea level, resulting in
// all points below sea level being negative.
func ModSeaLevel(min, max, q float64) Modify {
	delta := min + (max-min)*0.1
	//delta := quantile(h, q)
	return func(val float64) float64 {
		return val - delta
	}
}

// ModifyWithIndex is a function that modifies a value in a heightmap given
// its index and current value.
type ModifyWithIndex func(idx int, val float64) float64

// GetNeighbors returns all neighbor indices of an index on the heightmap.
type GetNeighbors func(idx int) []int

// GetHeight returns the height of a point on the heightmap given its index.
type GetHeight func(idx int) float64

// ModRelax applies a relaxation algorithm to the heightmap.
func ModRelax(n GetNeighbors, h GetHeight) ModifyWithIndex {
	return func(idx int, val float64) float64 {
		vals := []float64{val}
		for _, nb := range n(idx) {
			vals = append(vals, h(nb))
		}
		return CalcMean(vals)
	}
}
