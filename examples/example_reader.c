#include <stdio.h>
#include <stdint.h>
#include <inttypes.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>

typedef struct RockstarParticle {
  uint64_t ID;
  float X[3], V[3];
} RockstarParticle;

typedef struct GuppyHeader {
  uint64_t Version, Format;
  int64_t N, NTot;
  int64_t Span[3], Origin[3], TotalSpan[3];
  double Z, OmegaM, OmegaL, H100, L, Mass;
} GuppyHeader;

void FreadWithError(void *ptr, size_t elem_size, size_t elems,
					FILE *f, char *desc) {
  size_t n = fread(ptr, elem_size, elems, f);
  if (n != elems) {
	if (feof(f)) {
	  fprintf(stderr, "Failed to read %s with error: EOF reach after "
			  "%zu bytes instead of %zu.\n", desc, n, elem_size*elems);
	}

	fprintf(stderr, "Failed to read %s with error: %s\n",
			desc, strerror(errno));
	exit(1);
  }
}

int main() {
  // Run a guppy read process, tell it what file to read and tell it what
  // properties to get, in order.
  char *cmd = "../guppy read "
	"-file ../large_test_data/L125_sheet000_snap_100.gadget2.dat.gup "
	"-vars {RockstarParticle},x{0},v,id";

  FILE *pipe = popen(cmd, "r");

  // Read the Header
  GuppyHeader hd;
  FreadWithError(&hd, sizeof(hd), 1, pipe, "header");

  // Read the data
  RockstarParticle *part = calloc(sizeof(RockstarParticle), hd.N);
  float *x0 = calloc(sizeof(float), hd.N);
  float (*v)[3] = calloc(sizeof(float)*3, hd.N);
  uint64_t *id = calloc(sizeof(uint64_t), hd.N);

  FreadWithError(part, sizeof(struct RockstarParticle), hd.N, pipe,
				 "'{RockstarParticle}'");
  FreadWithError(x0, sizeof(float), hd.N, pipe, "'x{0}'");
  FreadWithError(v, sizeof(float)*3, hd.N, pipe, "'v'");
  FreadWithError(id, sizeof(uint64_t), hd.N, pipe, "id");
  
  pclose(pipe);

  for (int64_t i = 0; i < 8; i++) {
	printf("%9" PRIx64 " [%.4f %.4f %.4f] [%.4f %.4f %.4f] %.4f\n",
		   part[i].ID, part[i].X[0], part[i].X[1], part[i].X[2],
		   part[i].V[0], part[i].V[1], part[i].V[2], x0[i]);
  }
  
  return 0;
}
