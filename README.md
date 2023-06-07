# dbdiff
`dbdiff` is a databases comparison tool, allowing to compare schemas and data in databases of different types. Current version allows comparing databases:
- SQLite - SQLite, 
- PostgreSQL - PostgreSQL,
- SQLite - PostgreSQL,
- PostgreSQL - SQLite,
- as plain text files.

## Author
Yaroslav Grebnov

## Installation

**homebrew tap**
```shell
brew install dbdiff/tap/dbdiff
```

**homebrew**
```shell
brew install dbdiff
```

**go install**
```shell
go install github.com/ygrebnov/dbdiff
```

**manually**

Archives with pre-compiled binaries can be downloaded from [releases page](https://github.com/ygrebnov/dbdiff/releases). 

## Usage

**Example: SQLite - SQLite**
```shell
dbdiff sqlite:./d1.db sqlite:./d2.db
```

In the example above:
- `sqlite` is the database type,
- `./d1.db` and `./d2.db` are the paths to the databases files,
- `:` separates database type from the path.

**Example: SQLite - PostgreSQL**
```shell
dbdiff sqlite:./d1.db postgres:postgres://user:password@hostname:port/dbname?sslmode=disable
```

In this example, PostgreSQL database is identified as a combination of type `postgres` and connection URI `postgres://user:password@hostname:port/dbname?sslmode=disable` separated by colon. More information on PostgreSQL connection URIs can be found in the [official documentation](https://www.postgresql.org/docs/current/libpq-connect.html) (section 34.1.1.2).

**Example: comparing as files**
```shell
dbdiff -f ./d1.sql ./d2.sql
```

In this example, databases are compared as plain text files, line by line.

## Comparison results

Comparison results verbosity is configurable. By default, in case there are no differences found neither in schemas, nor in data, nothing is written to the console. In case there are some differences, they are output to the console in tabular format.

Output verbosity can be increased by specifying options: `-v` (level 1), `-vv` (level 2), or `-vvv` (level 3).

At level 1, all the compared tables are listed. For each table:
- if schemas are equal, outputs 'schema differences: none',
- if data is equal, outputs 'data differences: none'.

At level 2, the output is the same as at level 1, plus in case of equal schemas outputs the whole schema.

At level 3, the output is the same as at level 2, plus in case of equal data, outputs all the data.

**Example: level 2 verbosity**
```shell
dbdiff -vv sqlite:./d1.db sqlite:./d2.db
```
