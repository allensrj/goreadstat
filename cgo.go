package goreadstat

/*
#cgo CFLAGS: -I${SRCDIR}/readstat/src -std=c99 -DHAVE_ZLIB=1
#cgo darwin LDFLAGS: ${SRCDIR}/build/libreadstat.a -liconv -lz -lm
#cgo linux LDFLAGS: ${SRCDIR}/build/libreadstat.a -lz -lm

#include <readstat.h>
#include <stdlib.h>
#include "readstat_helpers.h"
*/
import "C"
