package main

import (
	"fmt"
	"math"
	"os"

	"gonum.org/v1/gonum/mat"

	"github.com/phil-mansfield/guppy/lib/catio"
	"github.com/phil-mansfield/guppy/lib/compress"
)

const (
	L = 62.5
	NTot = 1<<30
	Mp = 1.70e7 // Msun/h
	SimName = "Erebos_CBol_L63"
	XFileBase = "/data/mansfield/simulations/Erebos_CBol_L63/particles/guppy/dx_20/snapdir_100/sheet%d%d%d.gup"
	VFileBase = "/data/mansfield/simulations/Erebos_CBol_L63/particles/guppy/dv_20/snapdir_100/sheet%d%d%d.gup"
	BoundsName = "Erebos_CBol_L63.bounds.txt"
	ProfileDir = "profiles/delta_20_20"
	Snap = 100
)

var (
	VecHalo = [][3]float32{
		{39.09977, 21.61123, 50.9799 },
		{48.7807 , 44.43203, 45.31415},
		{35.27232, 23.53042, 32.002  },
		{29.84817, 18.79563, 35.643  },
		{42.34903, 21.92361, 51.54549},
		{40.02427, 53.96078, 34.52372},
		{48.42057, 19.49928, 53.84796},
		{ 5.85089,  6.55777, 19.45434},
		{ 3.83299,  9.02794,  7.4789 },
		{ 2.53586, 51.90359, 38.44151},
	}
	VelHalo = [][3]float32{
		{ 3.6555e+02, -3.7184e+02, -1.6995e+02},
		{ 5.5200e+01,  1.3279e+02, -1.9325e+02},
		{-2.9440e+01, -2.0304e+02,  1.9189e+02},
		{ 1.3958e+02,  2.6710e+01,  3.9000e-01},
		{ 5.0990e+01, -3.3153e+02, -1.5383e+02},
		{-3.1000e-01, -2.6570e+01,  2.2095e+02},
		{-2.4392e+02, -1.5097e+02, -1.7828e+02},
		{-4.0110e+01,  4.8040e+01,  3.9040e+01},
		{ 3.5740e+01, -1.5160e+01,  1.7210e+02},
		{ 4.5780e+01,  2.2764e+02, -6.9480e+01},
	}
	Rmax = float32(math.Pow(10, 0.5))
	RHalo = []float32{ Rmax, Rmax, Rmax, Rmax, Rmax,
		Rmax, Rmax, Rmax, Rmax, Rmax }
	MassHalo = []float32{ 5.742e+13, 5.916e+13, 6.081e+13, 6.248e+13, 9.282e+13,
		9.547e+13, 1.041e+14, 1.387e+14, 2.181e+14, 3.166e+14 }
)

type Bounds [2][3]float32

func FileBounds(boundsName string) []Bounds {
	file := catio.TextFile(boundsName)
	cols := file.ReadFloat32s([]int{1, 2, 3, 4, 5, 6})

	bounds := make([]Bounds, len(cols[0]))
	for i := range bounds {
		for dim := 0; dim < 3; dim++ {
			bounds[i][0][dim] = cols[dim][i]
			bounds[i][1][dim] = cols[dim + 3][i]
		}
	}

	return bounds
}

func (b1 Bounds) Intersect(b2 Bounds, L float32) bool {
	origin1, span1 := b1[0], b1[1]
	origin2, span2 := b2[0], b2[1]
    return intersect1d(origin1[0], span1[0], origin2[0], span2[0], L) &&
        intersect1d(origin1[1], span1[1], origin2[1], span2[1], L) &&
        intersect1d(origin1[2], span1[2], origin2[2], span2[2], L)
}

func intersect1d(x1, w1, x2, w2, L float32) bool {
    return oneWayIntersect(x1, w1, x2, L) ||
        oneWayIntersect(x2, w2, x1, L)
}

func oneWayIntersect(x1, w1, x2, L float32) bool {
    if x1 > x2 { x1 -= L }
    return x1 + w1 > x2
}

func BallToBounds(x [3]float32, r, L float32) Bounds {
	b := Bounds{ }

	for i := 0; i < 3; i++ {
		b[0][i] = x[i] - r
		if b[0][i] < 0 { b[0][i] += L }
		
		b[1][i] = 2*r
	}

	return b
}

func IntersectingFiles(vec [3]float32, r float32, bounds []Bounds) []int {
	halo := BallToBounds(vec, r, L)
	idx := []int{ }

	for i := range bounds {
		if halo.Intersect(bounds[i], L) {
			idx = append(idx, i)
		}
	}

	return idx
}

func PeriodicDisplacement(x1, x2 [3]float64, L float64) [3]float64 {
	out := [3]float64{ }
	for dim := 0; dim < 3; dim++ {
		dx := x1[dim] - x2[dim]
		if dx > L/2 { 
			dx -= L
		} else if dx < -L/2 {
			dx += L
		}
		out[dim] = dx
	}
	return out
}

func Displacement(x1, x2 [3]float64) [3]float64 {
	out := [3]float64{ }
	for dim := 0; dim < 3; dim++ {
		out[dim] = x1[dim] - x2[dim]
	}
	return out
}

func PeriodicR2(x1, x2 [3]float32, L float32) float32 {
	r2 := float32(0.0)

	for dim := 0; dim < 3; dim++ {
		dx := x1[dim] - x2[dim]
		if dx > L/2 { 
			dx -= L
		} else if dx < -L/2 {
			dx += L
		}

		r2 += dx*dx
	}	

	return r2
}

func FilterVec(
	x, v, fx, fv [][3]float32, hx [3]float32, hr, L float32,
) (fxOut, fvOut [][3]float32) {
	hr2 := hr*hr

	for i := range x {
		r2 := PeriodicR2(x[i], hx, L)
		if r2 <= hr2 {
			fx, fv = append(fx, x[i]), append(fv, v[i])
		}
	}

	return fx, fv
}

func CalculateProfiles(x, v [][3]float32, i int) {
	DensityProfile(x, v, i)
	CricularVelocityProfile(x, v, i)
	ShapeProfile(x, v, i)
	AngularVelocityProfile(x, v, i)
	//PericenterProfile(x, v, i)
	//ApocenterProfile(x, v, i)
}



// First bin is all the particles inside rMin.
func radialHist(
	nBins int, rMin, rMax float64, x, v [][3]float32, i int,
) (xHist, vHist [][][3]float64) {
	xh := VecHalo[i]

	logRMin, logRMax := math.Log10(rMin), math.Log10(rMax)
	dLogR := (logRMax - logRMin) / float64(nBins)

	xHist = make([][][3]float64, nBins + 1)
	vHist = make([][][3]float64, nBins + 1)

	for j := range x {
		r2 := PeriodicR2(x[j], xh, L)
		logR := math.Log10(float64(r2)) / 2

		ri := int(math.Floor((logR - logRMin) / dLogR))
		if ri >= nBins { continue }
		if ri < 0 { ri = -1 }

		xHist[ri+1] = append(xHist[ri+1], [3]float64{
			float64(x[j][0]), float64(x[j][1]), float64(x[j][2]),
		})
		vHist[ri+1] = append(vHist[ri+1], [3]float64{
			float64(v[j][0]), float64(v[j][1]), float64(v[j][2]),
		})
	}

	return xHist, vHist
}

func radialBinEdges(nBins int, rMin, rMax float64) []float64 {
	r := make([]float64, nBins + 1)
	logRMin, logRMax := math.Log10(rMin), math.Log10(rMax)
	dLogR := (logRMax - logRMin) / float64(nBins)
	for i := range r {
		logR := float64(i)*dLogR + logRMin 
		r[i] = math.Pow(10, logR)
	}

	return r
}

func radialBinCenters(nBins int, rMin, rMax float64) []float64 {
	edges := radialBinEdges(nBins, rMin, rMax)
	r := make([]float64, len(edges) - 1)
	for i := range r {
		logR := (math.Log10(edges[i]) + math.Log10(edges[i]))/2
		r[i] = math.Pow(10, logR)
	}

	return r
}

func PrintToFile(i int, varName, comment string, columns ...[]float64) {
	f, err := os.Create(fmt.Sprintf("%s/%s.%d.txt", ProfileDir, varName, i))
	if err != nil { panic(err.Error()) }
	defer f.Close()

	fmt.Fprintf(f, "# Halo %d:\n", i)
	fmt.Fprintf(f, "# Mvir = %.5g Msun/h", MassHalo[i])
	fmt.Fprintf(f, "# X = %.5f Mpc/h\n", VecHalo[i])
	fmt.Fprintf(f, "# V = %.5f km/s\n", VelHalo[i])
	fmt.Fprintf(f, "%s\n", comment)

	for i := range columns[0] {
		for j := range columns {
			fmt.Fprintf(f, "%9.4g ", columns[j][i])
		}
		fmt.Fprintf(f, "\n")
	}
}

func shellVolume(r1, r2 float64) float64 {
	return math.Abs(4*math.Pi/3 * (r2*r2*r2 - r1*r1*r1))
}

func DensityProfile(x, v [][3]float32, i int) {
	rMin, rMax := math.Pow(10, -3.5), math.Pow(10, 0.5)
	nBins := 40

	xHist, _ := radialHist(nBins, rMin, rMax, x, v, i)
	rEdges := radialBinEdges(nBins, rMin, rMax)
	rMids := radialBinCenters(nBins, rMin, rMax)

	avgNumDensity := NTot / float64(L*L*L)

	out := make([]float64, nBins)
	for j := range out {
		vol :=  shellVolume(rEdges[j], rEdges[j+1])
		numDensity := float64(len(xHist[j+1])) / vol
		out[j] = numDensity / avgNumDensity * rMids[j]*rMids[j]
	}

	comment := "# 0 - R (Mpc/h)\n # r^2 rho / rho_mean (Mpc/h)"
	PrintToFile(i, "rho", comment, rMids, out)
}

func vCirc(m, r float64) float64 {
	return math.Sqrt(m / r) * 6.558e-5
}

func CricularVelocityProfile(x, v [][3]float32, i int) {
	rMin, rMax := math.Pow(10, -3.5), math.Pow(10, 0.5)
	nBins := 200

	xHist, _ := radialHist(nBins, rMin, rMax, x, v, i)
	rEdges := radialBinEdges(nBins, rMin, rMax)

	out := make([]float64, nBins)

	cSum := len(xHist[0])
	for j := range out {
		cSum += len(xHist[j+1])
		cMass := Mp*float64(cSum)
		out[j] = vCirc(cMass, rEdges[j+1])
	}

	comment := "# 0 - R (Mpc/h)\n # v(<R) (km/s)"
	PrintToFile(i, "vcirc", comment, rEdges[1:], out)
}

func axisRatios(x [][3]float64, xh [3]float32) (ca, ba float64) {
	if len(x) < 4 { return -1, -1 }

	xh64 := [3]float64{ float64(xh[0]), float64(xh[1]), float64(xh[2]) }
	S := make([]float64, 9)

	for k := range x {
		dx := PeriodicDisplacement(x[k], xh64, L)
		r2 := dx[0]*dx[0] + dx[1]*dx[1] + dx[2]*dx[2]
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				S[i + 3*j] += dx[i]*dx[j]/r2
			}
		}
	}

	for i := range S { S[i] /= float64(len(x)) }
	Smat := mat.NewDense(3, 3, S)

	eig := &mat.Eigen{ }
	ok := eig.Factorize(Smat, mat.EigenRight)
	if !ok { panic(fmt.Sprintf("decomposition of %v failed", Smat)) }
	val := eig.Values(make([]complex128, 3))
	
	a2, b2, c2 := sort3(real(val[0]), real(val[1]), real(val[2]))
	return math.Sqrt(c2/a2), math.Sqrt(b2/a2)
}

func sort3(x, y, z float64) (l1, l2, l3 float64) {
	min, max := x, x
	if y > max {
		max = y
	} else if y < min {
		min = y
	}

	if z > max {
		max = z
	} else if z < min {
		min = z
	}

	return max, (x+y+z) - (min+max), min
}

func ShapeProfile(x, v [][3]float32, i int) {
	rMin, rMax := math.Pow(10, -3.5), math.Pow(10, 0.5)
	nBins := 40

	xHist, _ := radialHist(nBins, rMin, rMax, x, v, i)
	rMids := radialBinCenters(nBins, rMin, rMax)

	ca, ba := make([]float64, nBins), make([]float64, nBins)
	for j := range ca {
		ca[j], ba[j] = axisRatios(xHist[j+1], VecHalo[i])
	}

	comment := "# 0 - R (Mpc/h)\n# 1 - c/a\n# 2 - b/a"
	PrintToFile(i, "shape", comment, rMids, ca, ba)
}

func cross(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0],
	}
}

func norm(x [3]float64) float64 {
	return math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
}

func specificJ(x, v [][3]float64, xh, vh [3]float32) [3]float64 {
	if len(x) == 0 { return [3]float64{ } }

	J := [3]float64{ }

	xh64 := [3]float64{ float64(xh[0]), float64(xh[1]), float64(xh[2]) }
	vh64 := [3]float64{ float64(vh[0]), float64(vh[1]), float64(vh[2]) }

	for i := range x {
		dx, dv := PeriodicDisplacement(x[i], xh64, L), Displacement(v[i], vh64)
		Ji := cross(dx, dv)
		for dim := 0; dim < 3; dim++ { J[dim] += Ji[dim] }
	}

	for dim := 0; dim < 3; dim++ { J[dim] /= float64(len(x)) }

	return J
}

func AngularVelocityProfile(x, v [][3]float32, i int) {
	rMin, rMax := math.Pow(10, -3.5), math.Pow(10, 0.5)
	nBins := 40

	xHist, vHist := radialHist(nBins, rMin, rMax, x, v, i)
	rMids := radialBinCenters(nBins, rMin, rMax)
	
	out := make([]float64, nBins)

	cSum := len(xHist[0])
	for j := range out {
		if len(xHist[j+1]) == 0 {
			out[j] = -1 
			continue
		}

		cSum += len(xHist[j+1])
		cMass := Mp*float64(cSum)
		vc := vCirc(cMass, rMids[j])
		J := specificJ(xHist[j+1], vHist[j+1], VecHalo[i], VelHalo[i])
		out[j] = norm(J)/(rMids[j]*vc)
	}

	comment := `# 0 - R (Mpc/h)
# 1 - lambda = J / (R * Vcirc(<R)) (Mpc/h * km/s)^-1`
	PrintToFile(i, "l_bullock", comment, rMids, out)
}

func PericenterProfile(x, v [][3]float32, i int) {
	panic("NYI")
}

func ApocenterProfile(x, v [][3]float32, i int) {
	panic("NYI")
}

func getFileNames() (xNames, vNames []string) {
	for iz := 0; iz <= 7; iz++ {
		for iy := 0; iy <= 7; iy++ {
			for ix := 0; ix <= 7; ix++ {
				xNames = append(xNames, fmt.Sprintf(XFileBase, ix, iy, iz))
				vNames = append(vNames, fmt.Sprintf(VFileBase, ix, iy, iz))
			}
		}
	}

	return xNames, vNames
}

func AnalyzeHalo(iHalo int) {
	bounds := FileBounds(BoundsName)
	idx := IntersectingFiles(VecHalo[iHalo], RHalo[iHalo], bounds)

	xFileNames, vFileNames := getFileNames()

	n := 128*128*128

	xBuf := make([][3]float32, n)
	vBuf := make([][3]float32, n)
	x, v := [][3]float32{ }, [][3]float32{ }

	buf := compress.NewBuffer(0)
	midBuf := []byte{ }

	for _, i := range idx {
		// Read in x
		rdX, err := compress.NewReader(xFileNames[i], buf, midBuf)
		if err != nil { panic(err.Error()) }

		for dim := 0; dim < 3; dim++ {
			field, err := rdX.ReadField(fmt.Sprintf("x[%d]", dim))
			if err != nil { panic(err.Error()) }

			xi, ok := field.Data().([]float32)
			if !ok { panic("Impossible!") }

			for j := range xi { xBuf[j][dim] = xi[j] }
		}

		midBuf = rdX.ReuseMidBuf()
		rdX.Close()

		// Read in v
		rdV, err := compress.NewReader(vFileNames[i], buf, midBuf)
		if err != nil { panic(err.Error()) }

		for dim := 0; dim < 3; dim++ {
			field, err := rdV.ReadField(fmt.Sprintf("v[%d]", dim))
			if err != nil { panic(err.Error()) }

			vi, ok := field.Data().([]float32)
			if !ok { panic("Impossible!") }

			for j := range vBuf { vBuf[j][dim] = vi[j] }
		}

		midBuf = rdV.ReuseMidBuf()
		rdX.Close()

		x, v = FilterVec(xBuf, vBuf, x, v, VecHalo[iHalo], RHalo[iHalo], L)
	}


	CalculateProfiles(x, v, iHalo)
}
func main() {
	for iHalo := range VecHalo {
		AnalyzeHalo(iHalo)
	}
}
