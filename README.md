# sql2entity

A tool for generate c# entity class from sql file
## Install
```
go get -u github.com/nangcr/sql2entity
```
## Usage
Simply generate entity class:

```
sql2entity [entity name] <file>
```
To add prefix and suffix for output:

Create file named 'prefix' or 'suffix' and put them in same directory with your '.sql' file.