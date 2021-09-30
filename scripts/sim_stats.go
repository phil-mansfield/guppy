package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"log"
	"strconv"

	"gonum.org/v1/gonum/mat"

	"github.com/phil-mansfield/guppy/lib/catio"
	"github.com/phil-mansfield/guppy/lib/compress"
	"github.com/phil-mansfield/gravitree"
)

const (
	L = 62.5
	NTot = 1<<30
	Mp = 1.70e7 // Msun/h
	Eps = 1e-3 // Mpc/h
	SimName = "Erebos_CBol_L63"
	XFileBase = "/data/mansfield/simulations/Erebos_CBol_L63/particles/guppy/dx_%s/snapdir_100/sheet%d%d%d.gup"
	VFileBase = "/data/mansfield/simulations/Erebos_CBol_L63/particles/guppy/dv_%s/snapdir_100/sheet%d%d%d.gup"
	BoundsName = "Erebos_CBol_L63.bounds.txt"
	ProfileDir = "profiles/delta_%s_%s"
	Snap = 100

	G = 4.30091e-9 // Mpc / Msun (kms/)^2
)

var (
	AccStrX = "-1"
	AccStrV = "-1"
)

var (
	VecHalo = [][3]float32{ }
	VelHalo = [][3]float32{ }
	MassHalo = []float32{ }

	TrueCenter = [][3]float32{
		{24.646318, 50.388489, 40.691135},
		{2.943359, 25.294899, 43.519661},
		{7.378795, 62.006016, 56.563351},
		{1.100551, 26.487192, 43.755669},
		{3.140521, 19.414217, 18.770119},
		{14.061264, 25.592108, 12.812575},
		{31.813524, 20.182394, 34.111942},
		{23.294991, 22.222094, 30.768160},
		{7.783168, 3.619000, 53.702065},
		{5.867202, 12.762835, 16.456970},
		{5.791772, 6.827592, 19.124933},
		{5.980662, 44.681656, 23.707287},
		{17.437038, 18.260902, 37.362926},
		{36.061024, 50.672203, 10.101091},
		{40.321720, 54.610283, 34.308846},
		{41.934315, 21.636044, 51.683231},
		{52.467125, 49.394814, 58.740944},
		{48.242256, 7.875600, 42.000328},
		{37.455990, 34.805462, 6.164632},
		{4.307057, 8.208133, 7.663888},
		{50.695938, 48.622326, 55.069504},
		{10.721359, 37.308678, 24.777666},
		{5.999717, 53.566811, 34.915215},
		{0.943466, 56.876759, 49.666649},
		{18.209871, 17.591198, 2.345853},
		{36.439930, 59.208130, 49.026863},
		{2.962955, 8.589188, 7.916003},
		{4.008523, 61.911388, 34.385906},
		{7.778188, 62.209148, 57.454563},
		{62.369953, 39.413383, 41.809113},
		{2.046329, 55.506100, 35.526775},
		{59.247417, 29.100414, 44.971905},
		{2.196865, 55.654625, 35.674904},
		{37.892406, 16.701111, 49.943054},
		{40.043858, 53.965565, 34.506462},
		{2.185093, 51.061249, 38.743866},
	}
)

func MvirToRvirZ0(mvir float32) float64 {
	return 0.4459216 * math.Pow(float64(mvir)/1e13, 1.0/3)
}

type Bounds [2][3]float32

func LoadTargetHaloes(fname string) {
	file := catio.TextFile(fname)
	cols := file.ReadFloat32s([]int{0, 2, 3, 4, 5, 6, 7})
	mvir, x, y, z := cols[0], cols[1], cols[2], cols[3]
	vx, vy, vz := cols[4], cols[5], cols[6]

	n := len(x)
	VecHalo = make([][3]float32, n)
	VelHalo = make([][3]float32, n)
	MassHalo = make([]float32, n)

	for i := 0; i < n; i++ {
		VecHalo[i][0] = float32(x[i])
		VecHalo[i][1] = float32(y[i])
		VecHalo[i][2] = float32(z[i])
		VelHalo[i][0] = float32(vx[i])
		VelHalo[i][1] = float32(vy[i])
		VelHalo[i][2] = float32(vz[i])
		MassHalo[i] = float32(mvir[i])
	}
}

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
	//log.Printf("Analyzing halo %d", i)

	pe := PotentialEnergy(x)
	ke := KineticEnergy(VelHalo[i], v)

	xb, vb := [][3]float32{ }, [][3]float32{ }

	for j := range pe {
		if pe[j] + ke[j] < 0 {
			xb, vb = append(xb, x[j]), append(vb, v[j])
		}
	}

	DensityProfile(xb, vb, i)
	CricularVelocityProfile(xb, vb, i)
	ShapeProfile(xb, vb, i)
	AngularVelocityProfile(xb, vb, i)
	BoundFractionProfile(x, xb, i)
	EmulateMostBoundDistribution(pe, x, v, i)
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

func radialIdxHist(
	nBins int, rMin, rMax float64, x [][3]float32, i int,
) (idx [][]int) {
	xh := VecHalo[i]

	logRMin, logRMax := math.Log10(rMin), math.Log10(rMax)
	dLogR := (logRMax - logRMin) / float64(nBins)

	idx = make([][]int, nBins + 1)

	for j := range x {
		r2 := PeriodicR2(x[j], xh, L)
		logR := math.Log10(float64(r2)) / 2

		ri := int(math.Floor((logR - logRMin) / dLogR))
		if ri >= nBins { continue }
		if ri < 0 { ri = -1 }

		idx[ri+1] = append(idx[ri+1], j)
	}

	return idx
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
	dir := fmt.Sprintf(ProfileDir, AccStrX, AccStrV)
	f, err := os.Create(fmt.Sprintf("%s/%s.%d.txt", dir, varName, i))
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
	rVir := MvirToRvirZ0(MassHalo[i])
	rMin := math.Pow(10, -3.5)
	nBins := 40

	xHist, _ := radialHist(nBins, rMin, rVir, x, v, i)
	rEdges := radialBinEdges(nBins, rMin, rVir)
	rMids := radialBinCenters(nBins, rMin, rVir)

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
	rVir := MvirToRvirZ0(MassHalo[i])
	rMin := math.Pow(10, -3.5)
	nBins := 200

	xHist, _ := radialHist(nBins, rMin, rVir, x, v, i)
	rEdges := radialBinEdges(nBins, rMin, rVir)

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
	rVir := MvirToRvirZ0(MassHalo[i])
	rMin := math.Pow(10, -3.5)
	nBins := 20

	xHist, _ := radialHist(nBins, rMin, rVir, x, v, i)
	rMids := radialBinCenters(nBins, rMin, rVir)

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
	rVir := MvirToRvirZ0(MassHalo[i])
	rMin := math.Pow(10, -3.5)
	nBins := 20

	xHist, vHist := radialHist(nBins, rMin, rVir, x, v, i)
	rMids := radialBinCenters(nBins, rMin, rVir)
	
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

// Specific kinetic energy. In units of (km/s)^2
func KineticEnergy(vh [3]float32, v [][3]float32) []float64 {
	ke := make([]float64, len(v))

	for i := range v {
		for dim := 0; dim < 3; dim++ {
			dv := float64(vh[dim] - v[i][dim])
			ke[i] += dv*dv
		}
		ke[i] /= 2
	}

	return ke
}

func PotentialEnergy(x [][3]float32) []float64 {
	dx := make([][3]float64, len(x))
	x0 := [3]float64{ float64(x[0][0]), float64(x[0][1]), float64(x[0][2]) }
	for i, xx := range x {
		xi := [3]float64{ float64(xx[0]), float64(xx[1]), float64(xx[2]) }
		dx[i] = PeriodicDisplacement(xi, x0, L)
	}

	tree := gravitree.NewTree(dx)

	pe := make([]float64, len(x))
	tree.Potential(Eps, pe)
	for i := range pe { pe[i] *= G*Mp }

	return pe
}

func BoundFractionProfile(x, xb [][3]float32, i int) {
	rVir := MvirToRvirZ0(MassHalo[i])
	rMin := math.Pow(10, -3.5)
	nBins := 20

	idx := radialIdxHist(nBins, rMin, rVir, x, i)
	rMids := radialBinCenters(nBins, rMin, rVir)

	xHist, _ := radialHist(nBins, rMin, rVir, x, x, i)
	xbHist, _ := radialHist(nBins, rMin, rVir, xb, xb, i)

	out := make([]float64, nBins)
	for j := range out {
		if len(idx[j+1]) == 0 {
			out[j] = -1 
			continue
		}

		out[j] = float64(len(xbHist[j+1])) / float64(len(xHist[j+1]))
	}

	comment := `# 0 - R (Mpc/h)
# 1 - M_bound(<R) / M_total(<R)`
	PrintToFile(i, "f_bound", comment, rMids, out)
}

func iMin(x []float64) int {
	im := 0
	for i := range x {
		if x[i] < x[im] { im = i }
	}
	return im
}

func EmulateMostBoundDistribution(
	pe []float64, x, v [][3]float32, i int,
) {
	dv64, _ := strconv.ParseFloat(AccStrV, 64)
	dv := float32(dv64)

	vh, xh := VelHalo[i], TrueCenter[i]
	energy := make([]float64, len(v))

	logRMin, logRMax := -4.0, 0.0
	bins := 80
	dlogR := (logRMax - logRMin) / float64(bins)

	hist := make([]float64, bins+1)
	histR := make([]float64, bins+1)
	for j := 0; j < bins; j++ {
		histR[j+1] = logRMin + dlogR*(float64(j) + 0.5)
	}

	for trial := 0; trial < 1000; trial++ {
		vTrial := EmulatePixelation(v, dv)
		ke := KineticEnergy(vh, vTrial)

		for j := range energy {
			energy[j] = ke[j] + pe[j]
		}

		im := iMin(energy)

		r2 := PeriodicR2(x[im], xh, L)
		logR := math.Log10(math.Sqrt(float64(r2)))

		if logR <= logRMin {
			hist[0]++
			continue
		}

		idx := int(math.Floor((logR - logRMin) / dlogR))
		if idx < 0 || idx >= bins { continue }
		hist[idx+1]++

		if trial == 0 {
			fmt.Println("# r, phi + ke")
			for j := range energy {
				r2 := PeriodicR2(x[j], xh, L)
				fmt.Printf("%7.4f %7.4f\n",
					math.Log10(math.Sqrt(float64(r2))), energy[j],
				)
			}
		}
	}

	//fmt.Printf("%.0f\n", hist)
	//fmt.Printf("%.3f\n", histR)

	comment := `# 0 - R (Mpc/h)
# 1 - M_bound(<R) / M_total(<R)`
	PrintToFile(i, "r_min_distr", comment, histR, hist)
}


func EmulatePixelation(x [][3]float32, dx float32) [][3]float32 {
	out := make([][3]float32, len(x))
	for i :=  range x {
		for dim := 0; dim < 3; dim++ {
			out[i][dim] =float32(float64(dx)*(math.Floor(float64(x[i][dim]/dx)) +
				rand.Float64()))
		}
	}
	return out
}

func getFileNames() (xNames, vNames []string) {
	for iz := 0; iz <= 7; iz++ {
		for iy := 0; iy <= 7; iy++ {
			for ix := 0; ix <= 7; ix++ {
				xName := fmt.Sprintf(XFileBase, AccStrX, ix, iy, iz)
				vName := fmt.Sprintf(VFileBase, AccStrV, ix, iy, iz)
				xNames = append(xNames, xName)
				vNames = append(vNames, vName)
			}
		}
	}

	return xNames, vNames
}

func AnalyzeHalo(iHalo int) {	
	mult := 1.25

	bounds := FileBounds(BoundsName)
	rVir := MvirToRvirZ0(MassHalo[iHalo])
	idx := IntersectingFiles(VecHalo[iHalo], float32(rVir*mult), bounds)

	log.Printf("Reading %d files for halo %d", len(idx), iHalo)

	xFileNames, vFileNames := getFileNames()

	n := 128*128*128

	xBuf := make([][3]float32, n)
	vBuf := make([][3]float32, n)
	x, v := [][3]float32{ }, [][3]float32{ }

	buf := compress.NewBuffer(0)
	midBuf := []byte{ }

	for ii, i := range idx {
		if ii != 0 && ii % 5 == 0 {
			//log.Printf("%d fields read", ii)
		}
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

			for j := range vBuf {
				vBuf[j][dim] = vi[j]
			}
		}

		midBuf = rdV.ReuseMidBuf()
		rdX.Close()

		x, v = FilterVec(xBuf, vBuf, x, v, VecHalo[iHalo],float32(rVir*mult), L)
	}


	CalculateProfiles(x, v, iHalo)
}


func main() {
	AccStrX, AccStrV = os.Args[1], os.Args[2]
	LoadTargetHaloes("profiles/target_haloes.txt")
	
	for iHalo := range MassHalo {
		if iHalo != 7 { continue }
		AnalyzeHalo(iHalo)
	}
}
