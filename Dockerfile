FROM golang:1.21

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum .
RUN go mod download

# Copy sources
COPY src etc htdocs ./

# Build
RUN go build -o mod-reporting ./...

EXPOSE 12369

# Run
ENV LOGCAT=listen,op,curl,status,response,db
ENV OKAPI_URL=https://folio-snapshot-okapi.dev.folio.org
ENV OKAPI_TENANT=diku
ENV OKAPI_USER=diku_admin
ENV OKAPI_PW=swordfish
ENV REPORTING_DB_URL=postgres://id-test-metadb.folio.indexdata.com:5432/metadb_indexdata_test
ENV REPORTING_DB_USER=miketaylor
ENV REPORTING_DB_PASS=swordfish
CMD ["./mod-reporting", "config.json"]
