# App

An opinionated approach to build golang application.

Featuring:
- [x] CLI Application (using [cobra](https://github.com/spf13/cobra))
- [x] Configuration files (using [viper](https://github.com/spf13/viper))
    - [ ] dotEnv support
    - [ ] Consul suuport
- [x] OpenTelemetry
- [ ] HTTP Server
    - [ ] Standard library
    - [ ] Gin
- [ ] HTTP Client
- [ ] Data Storages
    - [ ] Postgres using [lib/pq](https://github.com/lib/pq)
    - [x] Postgres using [pgx](https://github.com/jackc/pgx)
    - [ ] Mysql
    - [ ] MSSQL
    - [ ] SQLite
    - [ ] Oracle
    - [ ] Redis
    - [ ] ElasticSearch
    - [ ] MongoDB
    - [ ] S3 Storage
- [ ] DB Migration
- [ ] Mail
- [ ] Generator
    - [ ] SQLC
    - [ ] OpenAPI

## Application Flow

- app.Run()
  - Init Config (if config file defined)
  - Init Telemetry (if defined in config)
  - Execute Command
  - Shutdown