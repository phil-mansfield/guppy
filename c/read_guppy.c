#include <stdio.h>
#include <stdlib.h>
#include <inttypes.h>
#include "read_guppy.h"
#include "go_wrapper.h"

GuppyHeader *GuppyReadHeader(char *fileName) {
	return  ReadHeader(fileName);
}

void GuppyFreeHeader(GuppyHeader *hd) {
	free(hd->OriginalHeader);
	for (int i = 0; i < hd->NVars; i++) {
		free(hd->Names[i]);
		free(hd->Types[i]);
	}
	free(hd->Names);
	free(hd->Types);
}


void GuppyPrintHeader(GuppyHeader *hd) {
	printf("OriginalHeader:\n");
	printf("Names:\n");
	printf("    [");
	for (int64_t i = 0; i < hd->NVars; i++) {
		printf("'%s', ", hd->Names[i]);
	}
	printf("'%s']\n", hd->Types[hd->NVars - 1]);
	printf("Types:\n");
	printf("    [");
	for (int64_t i = 0; i < hd->NVars; i++) {
		printf("'%s', ", hd->Types[i]);
	}
	printf("'%s']\n", hd->Types[hd->NVars - 1]);
	printf("Span:\n    [%"PRId64", %"PRId64", %"PRId64"]\n",
		hd->Span[0], hd->Span[1], hd->Span[2]);
	printf("Z:\n    %.6f\n", hd->Z);
	printf("OmegaM:\n    %.6f\n", hd->OmegaM);
	printf("L:\n    %.6f\n", hd->L);
	printf("H100:\n    %.6f\n", hd->H100);
	printf("Mass:\n    %.6g\n", hd->Mass);
}

void GuppyReadVar(char *fileName, char *varName, int workerID, void *out) {
	ReadVar(fileName, varName, workerID, out);
}

void GuppyInitWorkers(int n) {
	InitWorkers(n);
}

void _PrintGuppyArrays(float (*x)[3], float (*v)[3], float *x0, uint64_t *id) {
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
	GuppyHeader *hd = GuppyReadHeader(fileName);
	GuppyPrintHeader(hd);

	GuppyInitWorkers(2);

	float (*x)[3] = calloc(hd->N, sizeof(*x));
	float (*v)[3] = calloc(hd->N, sizeof(*v));
	float *x0 = calloc(hd->N, sizeof(*x0));
	uint64_t *id = calloc(hd->N, sizeof(*id)); 

	GuppyReadVar(fileName, "x", 0, x);
	GuppyReadVar(fileName, "v", 1, v);
	GuppyReadVar(fileName, "x[0]", 0, x0);
	GuppyReadVar(fileName, "id", 1, id);

	_PrintGuppyArrays(x, v, x0, id);

	free(x);
	free(v);
	free(x0);
	free(id);

	GuppyFreeHeader(hd);
}