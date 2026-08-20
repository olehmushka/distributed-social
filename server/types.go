package server

type Addr string

type Name string

// DBDSN is this service's own Postgres connection string. Each service
// owns its data and never reaches into another service's database.
type DBDSN string

// NatsURL is the shared event bus every service connects to.
type NatsURL string
