# Backup and Restore

The data stored by Apicurio Registry should be backed up regularly. While the procedure is essentially simple, there are several ways to do it depending on the type of persistence backend and where it's running.

## PostgreSQL

## SQL Dump

[SQL Dump](https://www.postgresql.org/docs/12/backup-dump.html) is a simple procedure that works with any PostgreSQL installation. It uses the [pg_dump](https://www.postgresql.org/docs/12/app-pgdump.html) utility program to generate a file with SQL commands that can be used to recreate the database in the same state as it was at the time of the dump.

### Creating the Dump

pg_dump writes its result to the standard output and a common approach is to redirect the output to a file.

```
$ pg_dump dbname > dumpfile
```

pg_dump is a regular PostgreSQL client application, therefore it can be executed from any remote host that has access to the database. Like any other client, the operations it can perform are limited to the user permissions.

The database server pg_dump should connnect to is specified by the command line options `-h host` and `-p port`. Client authentication can be done in many ways as described in [Chapter 20](https://www.postgresql.org/docs/12/client-authentication.html) of the PostgreSQL documentation.

Large dump files can be reduced using a compression tool, such as gzip.

```
$ pg_dump dbname | gzip > filename.gz
```

### Restoring the Dump

Dump files created by pg_dump aren't restored with the same program, they are restored using `psql`. Also, it's important to note the SQL dump doesn't include a command to create the database, so that needs to happen as a separate step before executing `psql`.

The usual commands to restore a SQL dump are:

```
$ createdb -T template0 dbname
$ psql dbname < dumpfile
```

Before restoring an SQL dump, all the users who own objects or were granted permissions on objects in the dumped database must already exist.

After restoring a backup, it is wise to run [ANALYZE](https://www.postgresql.org/docs/12/sql-analyze.html) on each database so the query optimizer has useful statistics.
