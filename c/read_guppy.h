#ifndef _READ_GUPPY_H
#define _READ_GUPPY_H

#include <stdint.h>

// Guppy
typedef struct Guppy_Header {
	// OriginalHeader is the original header of the one of the original
	// simulation files. OriginalHeaderLength is the length of that header in
	// bytes.
	char *OriginalHeader;
	int64_t OriginalHeaderLength;
	// Names gives the names of all the variables stored in the file.
	// Types give the types of these variables. "u32"/"u64" give 32-bit and
	// 64-bit unsigned integers, respectively, and "f32"/"f64" give 32-bit
	// and 64-bit floats, respectively. NVars is the number of variables
	// stored in the file.
	char **Names, **Types;
	int64_t NVars;
	// N and NTot give the number of particles in the file and in the
	// total simulation, respectively.
	int64_t N, NTot;
	// Span gives the dimensions of the slab of particles in the file.
	// Span[0], Span[1], and Span[2] are the x-, y-, and z- dimensions.
	int64_t Span[3];
	// Z, OmegaM, H100, L, and Mass give the redshift, Omega_m,
	// H0 / (100 km/s/Mpc), box width in comoving Mpc/h, and particle
	// mass in Msun/h, respectively.
	double Z, OmegaM, OmegaL, H100, L, Mass;
} Guppy_Header;

// Guppy_RockstarParitcle has the same structure as the particles used 
// internally by Rockstar. Arrays of Guppy_RockstarParticle can be used
// so oOckstar doesn't need to make unneccessary heap allocations. 
typedef struct Guppy_RockstarParticle {
	uint64_t ID;
	float X[3], V[3];
} Guppy_RockstarParticle;

// Guppy_ReadHeader returns the header of a given file.
Guppy_Header *Guppy_ReadHeader(char *fileName);

// Guppy_FreeHeader frees the a Guppy_Header.
void Guppy_FreeHeader(Guppy_Header *hd);

// Guppy_PrintHeader prints a Guppy_Header.
void Guppy_PrintHeader(Guppy_Header *hd);

// Guppy_ReadVar reads a variable with a given name from a given file. If
// you and to use one of the pre-allocated workers, you should give the
// integer ID of that workers (i.e. in the range [0, n). ReadVar uses
// mutexes to make sure that same worker isn't being used simultaneously,
// so feel free to throw a zillion threads at the same worker. If you
// don't care about heap space, just set workerID to -1. If you want guppy
// to try to automatically allocate workers to the task, use workerID=-2. The
// last argument is a buffer with length Header.N where the variable will be
// written to.
//
// For vector quantities, you can either load each component one by one
// (e.g. "x[0]", "x[1]", etc.) and supply a []float32 or []float64 buffer,
// or you can get the full vector (e.g. "x") and supply a [][3]float32 or
// [][3]float64.
//
// The variable "id" is implicitly contained in every .gup file and can be
// read into a []uint64 array.
//
// If the buffer has the name "[RockstarParticle]" and type
// []Guppy_RockstarParticle, the fields "x[0]", "x[1]", "x[2]" will be read
// into the X field, "v[0]", "v[1]", and "v[2]" into the V field and "id"
// into the ID field.
void Guppy_ReadVar(char *fileName, char *varName, int workerID, void *out);

// Guppy_InitWorkers allocates memory-managed space for n workers which can
// be run simultaneously by different threads.
void Guppy_InitWorkers(int n);

#endif // _READ_GUPPY_H