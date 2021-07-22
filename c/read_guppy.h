#ifndef _READ_GUPPY_H
#define _READ_GUPPY_H

#include <stdint.h>

typedef struct GuppyHeader {
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
	double Z, OmegaM, H100, L, Mass;
} GuppyHeader;

GuppyHeader *GuppyReadHeader(char *fileName);

void GuppyReadVar(char *fileName, char *varName, int workerID, void *out);

void GuppyInitWorkers(int n);

#endif // _READ_GUPPY_H