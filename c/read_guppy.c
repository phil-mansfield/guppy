#include <stdlib.h>
#include <stdio.h>
#include <inttypes.h>
#include "read_guppy.h"
#include "guppy_wrapper.h"

Guppy_Header *Guppy_ReadHeader(char *fileName) {
	return ReadHeader(fileName);
}

void Guppy_FreeHeader(Guppy_Header *hd) {
	free(hd->OriginalHeader);
	for (int i = 0; i < hd->NVars; i++) {
		free(hd->Names[i]);
		free(hd->Types[i]);
	}
	free(hd->Names);
	free(hd->Types);
	free(hd->Sizes);
}


void Guppy_PrintHeader(Guppy_Header *hd) {
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
	printf("Sizes:\n");
	printf("    [");
	for (int64_t i = 0; i < hd->NVars; i++) {
		printf("'%"PRId64"', ", hd->Sizes[i]);
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

void Guppy_ReadVar(char *fileName, char *varName, int workerID, void *out) {
	ReadVar(fileName, varName, workerID, out);
}

void Guppy_InitWorkers(int n) {
	InitWorkers(n);
}
