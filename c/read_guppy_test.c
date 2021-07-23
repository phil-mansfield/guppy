#include <stdlib.h>
#include <stdio.h>
#include <inttypes.h>
#include "read_guppy.h"

void PrintGuppyArrays(float (*x)[3], float (*v)[3], float *x0, uint64_t *id) {
	printf("x:\n[\n");
	for (int i = 0; i < 5; i++)
		printf("     [%7.4f %7.4f %7.4f]\n", x[i][0], x[i][1], x[i][2]);
	printf("]\n\n");

	printf("v:\n[\n");
	for (int i = 0; i < 5; i++)
		printf("     [%9.4f %9.4f %9.4f]\n", v[i][0], v[i][1], v[i][2]);
	printf("]\n\n");

	printf("x0:\n    [");
	for (int i = 0; i < 5; i++)
		printf("%7.4f ", x0[i]);
	printf("]\n\n");

	printf("id:\n    [");
	for (int i = 0; i < 5; i++)
		printf("%"PRIu64" ", id[i]);
	printf("]\n");
}

int main() {
	char *fileName = "../large_test_data/large_test.gup";
	Guppy_Header *hd = Guppy_ReadHeader(fileName);
	Guppy_PrintHeader(hd);

	Guppy_InitWorkers(2);

	float (*x)[3] = calloc(hd->N, sizeof(*x));
	float (*v)[3] = calloc(hd->N, sizeof(*v));
	float *x0 = calloc(hd->N, sizeof(*x0));
	uint64_t *id = calloc(hd->N, sizeof(*id)); 

	Guppy_ReadVar(fileName, "x", 0, x);
	Guppy_ReadVar(fileName, "v", 1, v);
	Guppy_ReadVar(fileName, "x[0]", 0, x0);
	Guppy_ReadVar(fileName, "id", 1, id);

	PrintGuppyArrays(x, v, x0, id);

	free(x);
	free(v);
	free(x0);
	free(id);

	Guppy_FreeHeader(hd);
}
