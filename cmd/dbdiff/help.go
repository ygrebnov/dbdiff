package main

const usage = "Usage: dbdiff [-f] [-h] [-v] [-vv] [-vvv] database1 database2"
const help = "NAME\n\tdbdiff - compare databases\n\n" +
	"SYNOPSIS\n\tdbdiff [-f] [-h] [-v] [-vv] [-vvv] database1 database2\n\n" +
	"DESCRIPTION\n\tdbdiff compares two given databases. Comparison is a three stage process:\n" +
	"\t1) each table is checked for presence in both databases,\n" +
	"\t2) schemas are compared for each table in two databases,\n" +
	"\t3) data is compared for each table in two databases.\n\n" +
	"\tdbdiff can compare SQLite and PostgreSQL databases, even if the two compared databases are of " +
	"different types.\n" +
	"\tComparison results output verbosity level is configurable.\n" +
	"\tWith -f option specified, the databases are compared as files, line by line.\n\n" +
	"The following options are available:\n\n" +
	"\t-f\t\tCompare databases as files. Files are compared line by line. " +
	"Recommended for comparing sql files.\n\n" +
	"\t-h\t\tDisplay this help and exit.\n\n" +
	"\t-version\t\tDisplay version number and build timestamp.\n\n" +
	"\t-v\t\tOutput databases comparison results at verbosity level 1.\n\n" +
	"\t-vv\t\tOutput databases comparison results at verbosity level 2.\n\n" +
	"\t-vvv\t\tOutput databases comparison results at verbosity level 3.\n\n" +
	"Identifying databases\n" +
	"\tBy default (without -f option), database identification string has format: type:URI.\n" +
	"\tType can have either \"sqlite\" or \"postgres\" value.\n" +
	"\tFor an SQLite database, the URI is a path to the data file. In case the specified path is not " +
	"an existing file, an exception is raised.\n" +
	"\tFor a PostgreSQL database, the URI is composed as specified in the official documentation.\n\n" +
	"\t\tExample: dbdiff \"sqlite:data.db\" \"postgres:postgres://user:password@hostname:port/dbname?sslmode=disable\"\n\n" +
	"Output verbosity level\n" +
	"\tThere are four comparison results output verbosity levels:\n" +
	"\t0 - outputs only the differences,\n" +
	"\t1 - always lists all the compared tables. For each table:\n" +
	"\t\t- if schemas are equal, outputs \"schema differences: none\",\n" +
	"\t\t- if data is equal, outputs \"data differences: none\",\n" +
	"\t2 - same as level 1, but in case of equal schemas outputs the whole schema,\n" +
	"\t3 - same as level 2, but in case of equal data, outputs all the data."
