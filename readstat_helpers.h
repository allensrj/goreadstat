#ifndef READSTAT_HELPERS_H
#define READSTAT_HELPERS_H

#include "readstat.h"

extern int goMetadataHandler(readstat_metadata_t *metadata, void *ctx);
extern int goVariableHandler(int index, readstat_variable_t *variable,
                             const char *val_labels, void *ctx);
extern int goHandleValueDouble(int obs_index, int var_index, double value, void *ctx);
extern int goHandleValueString(int obs_index, int var_index, const char *value, void *ctx);
extern int goHandleValueMissing(int obs_index, int var_index, int tag, void *ctx);
extern int goHandleValueLabel(const char *label_set, int value_type,
                              double num_key, const char *str_key,
                              const char *label, void *ctx);
extern long long goWriteBytes(const void *data, size_t len, void *ctx);

int c_value_handler(int obs_index, readstat_variable_t *variable,
                    readstat_value_t value, void *ctx);
int c_value_label_handler(const char *val_labels, readstat_value_t value,
                          const char *label, void *ctx);
readstat_error_t setup_read_handlers(readstat_parser_t *parser);
ssize_t c_data_writer(const void *data, size_t len, void *ctx);

#endif
