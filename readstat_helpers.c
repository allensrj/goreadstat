#include "readstat.h"
#include "readstat_helpers.h"

int c_value_handler(int obs_index, readstat_variable_t *variable,
                    readstat_value_t value, void *ctx) {
    int var_index = readstat_variable_get_index(variable);
    readstat_type_t type = readstat_value_type(value);

    if (readstat_value_is_system_missing(value)) {
        return goHandleValueMissing(obs_index, var_index, 0, ctx);
    }
    if (readstat_value_is_tagged_missing(value)) {
        return goHandleValueMissing(obs_index, var_index, (int)readstat_value_tag(value), ctx);
    }

    switch (type) {
    case READSTAT_TYPE_STRING:
    case READSTAT_TYPE_STRING_REF:
        return goHandleValueString(obs_index, var_index,
                                   readstat_string_value(value), ctx);
    case READSTAT_TYPE_INT8:
        return goHandleValueDouble(obs_index, var_index,
                                   (double)readstat_int8_value(value), ctx);
    case READSTAT_TYPE_INT16:
        return goHandleValueDouble(obs_index, var_index,
                                   (double)readstat_int16_value(value), ctx);
    case READSTAT_TYPE_INT32:
        return goHandleValueDouble(obs_index, var_index,
                                   (double)readstat_int32_value(value), ctx);
    case READSTAT_TYPE_FLOAT:
        return goHandleValueDouble(obs_index, var_index,
                                   (double)readstat_float_value(value), ctx);
    case READSTAT_TYPE_DOUBLE:
        return goHandleValueDouble(obs_index, var_index,
                                   readstat_double_value(value), ctx);
    }
    return READSTAT_HANDLER_OK;
}

int c_value_label_handler(const char *val_labels, readstat_value_t value,
                          const char *label, void *ctx) {
    readstat_type_t type = readstat_value_type(value);
    switch (type) {
    case READSTAT_TYPE_STRING:
    case READSTAT_TYPE_STRING_REF:
        return goHandleValueLabel(val_labels, (int)type, 0,
                                  readstat_string_value(value), label, ctx);
    case READSTAT_TYPE_INT32:
        return goHandleValueLabel(val_labels, (int)type,
                                  (double)readstat_int32_value(value), NULL, label, ctx);
    default:
        return goHandleValueLabel(val_labels, (int)type,
                                  readstat_double_value(value), NULL, label, ctx);
    }
}

readstat_error_t setup_read_handlers(readstat_parser_t *parser) {
    readstat_error_t err = READSTAT_OK;
    if ((err = readstat_set_metadata_handler(parser, goMetadataHandler)) != READSTAT_OK) return err;
    if ((err = readstat_set_variable_handler(parser, goVariableHandler)) != READSTAT_OK) return err;
    if ((err = readstat_set_value_handler(parser, c_value_handler)) != READSTAT_OK) return err;
    if ((err = readstat_set_value_label_handler(parser, c_value_label_handler)) != READSTAT_OK) return err;
    return READSTAT_OK;
}

ssize_t c_data_writer(const void *data, size_t len, void *ctx) {
    return (ssize_t)goWriteBytes(data, len, ctx);
}
