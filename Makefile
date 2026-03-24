READSTAT_SRC = readstat/src

CC = gcc
AR = ar
CFLAGS = -std=c99 -O2 -Wall -I$(READSTAT_SRC) -I$(READSTAT_SRC)/sas -I$(READSTAT_SRC)/spss -I$(READSTAT_SRC)/stata -I$(READSTAT_SRC)/txt -DHAVE_ZLIB=1
BUILDDIR = build

SRCS = \
	$(READSTAT_SRC)/CKHashTable.c \
	$(READSTAT_SRC)/readstat_bits.c \
	$(READSTAT_SRC)/readstat_convert.c \
	$(READSTAT_SRC)/readstat_error.c \
	$(READSTAT_SRC)/readstat_io_unistd.c \
	$(READSTAT_SRC)/readstat_malloc.c \
	$(READSTAT_SRC)/readstat_metadata.c \
	$(READSTAT_SRC)/readstat_parser.c \
	$(READSTAT_SRC)/readstat_value.c \
	$(READSTAT_SRC)/readstat_variable.c \
	$(READSTAT_SRC)/readstat_writer.c \
	$(READSTAT_SRC)/sas/ieee.c \
	$(READSTAT_SRC)/sas/readstat_sas.c \
	$(READSTAT_SRC)/sas/readstat_sas7bcat_read.c \
	$(READSTAT_SRC)/sas/readstat_sas7bcat_write.c \
	$(READSTAT_SRC)/sas/readstat_sas7bdat_read.c \
	$(READSTAT_SRC)/sas/readstat_sas7bdat_write.c \
	$(READSTAT_SRC)/sas/readstat_sas_rle.c \
	$(READSTAT_SRC)/sas/readstat_xport.c \
	$(READSTAT_SRC)/sas/readstat_xport_read.c \
	$(READSTAT_SRC)/sas/readstat_xport_write.c \
	$(READSTAT_SRC)/sas/readstat_xport_parse_format.c \
	$(READSTAT_SRC)/spss/readstat_por.c \
	$(READSTAT_SRC)/spss/readstat_por_parse.c \
	$(READSTAT_SRC)/spss/readstat_por_read.c \
	$(READSTAT_SRC)/spss/readstat_por_write.c \
	$(READSTAT_SRC)/spss/readstat_sav.c \
	$(READSTAT_SRC)/spss/readstat_sav_compress.c \
	$(READSTAT_SRC)/spss/readstat_sav_parse.c \
	$(READSTAT_SRC)/spss/readstat_sav_parse_timestamp.c \
	$(READSTAT_SRC)/spss/readstat_sav_parse_mr_name.c \
	$(READSTAT_SRC)/spss/readstat_sav_read.c \
	$(READSTAT_SRC)/spss/readstat_sav_write.c \
	$(READSTAT_SRC)/spss/readstat_spss.c \
	$(READSTAT_SRC)/spss/readstat_spss_parse.c \
	$(READSTAT_SRC)/spss/readstat_zsav_compress.c \
	$(READSTAT_SRC)/spss/readstat_zsav_read.c \
	$(READSTAT_SRC)/spss/readstat_zsav_write.c \
	$(READSTAT_SRC)/stata/readstat_dta.c \
	$(READSTAT_SRC)/stata/readstat_dta_parse_timestamp.c \
	$(READSTAT_SRC)/stata/readstat_dta_read.c \
	$(READSTAT_SRC)/stata/readstat_dta_write.c \
	$(READSTAT_SRC)/txt/commands_util.c \
	$(READSTAT_SRC)/txt/readstat_copy.c \
	$(READSTAT_SRC)/txt/readstat_sas_commands_read.c \
	$(READSTAT_SRC)/txt/readstat_spss_commands_read.c \
	$(READSTAT_SRC)/txt/readstat_schema.c \
	$(READSTAT_SRC)/txt/readstat_stata_dictionary_read.c \
	$(READSTAT_SRC)/txt/readstat_txt_read.c

OBJS = $(patsubst $(READSTAT_SRC)/%.c, $(BUILDDIR)/%.o, $(SRCS))

.PHONY: all clean test

all: $(BUILDDIR)/libreadstat.a

$(BUILDDIR)/libreadstat.a: $(OBJS)
	$(AR) rcs $@ $^

$(BUILDDIR)/%.o: $(READSTAT_SRC)/%.c
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -rf $(BUILDDIR)

test: $(BUILDDIR)/libreadstat.a
	go test -v -count=1 ./...
