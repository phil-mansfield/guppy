package particles

// IDOrder is an interface for mapping paritcles IDs to their 3D index into the
// simulation grid. This interface also supports mutli-reoslution simulations
// which are split up into "levels" of fixed resolution.
type IDOrder interface {
	// IDToIndex converts an ID to its 3-index equivalent in the level's grid.
	// It also returns the resolution level of the particle.
	IDToIndex(id uint64) (idx [3]int, level int)
	// IndexToID converts a 3-index within a level to its ID.
	IndexToID(i [3]int, level int) uint64

	// Levels returns the number of levels that this system of IDs uses.
	Levels() int
	// LevelOrigin returns a 3-index to the origin of a given level in units of
	// that resolution level.
	LevelOrigin(level int) [3]int
	// LevelSpan returns the 3-index representing the span of a given level in
	// units of that resolution level.
	LevelSpan(level int) [3]int
}

// Type assertions
var (
	_ IDOrder = &ZMajorUnigrid{ }
)

// ZMajorUnigrid is the IDOrder of a z-major uniform-mass grid. This is the
// ordering used by, e.g., 2LPTic and many other codes. See the IDOrder
// interface for documentation of the methods.
type ZMajorUnigrid struct {
	n int
	n64 uint64
}

// NewZMajorUnigrid returns a z-major uniform density grid with width n on each
// side.
func NewZMajorUnigrid(n int) *ZMajorUnigrid {
	return &ZMajorUnigrid{ n, uint64(n) }
}

func (g *ZMajorUnigrid) IDToIndex(id uint64) (idx [3]int, level int) {
	return [3]int{
		int(id / (g.n64 * g.n64)),
		int((id / g.n64) % g.n64),
		int(id % g.n64),
	}, 0
}

func (g *ZMajorUnigrid) IDToLevel(id uint64) int { return 0 }

func (g *ZMajorUnigrid) IndexToID(i [3]int, level int) uint64 {
	return uint64(i[2] + i[1]*g.n + i[0]*g.n*g.n)
}


func (g *ZMajorUnigrid) Levels() int { return 1 }

func (g *ZMajorUnigrid) LevelOrigin(level int) [3]int {
	return [3]int{ 0, 0, 0 }
}

func (g *ZMajorUnigrid) LevelSpan(level int) [3]int {
	return [3]int{ g.n, g.n, g.n }
}

