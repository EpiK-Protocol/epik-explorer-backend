module github.com/EpiK-Protocol/epik-explorer-backend

go 1.14

require (
	github.com/EpiK-Protocol/go-epik v0.4.2-0.20201026070559-b048e96c9753
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/filecoin-project/go-address v0.0.2-0.20200504173055-8b6f2fb2b3ef
	github.com/filecoin-project/go-jsonrpc v0.1.1-0.20200602181149-522144ab4e24
	github.com/filecoin-project/specs-actors v0.6.2-0.20200724193152-534b25bdca30
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/snappy v0.0.2-0.20200707131729-196ae77b8a26 // indirect
	github.com/ipfs/go-cid v0.0.6
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/shirou/gopsutil v2.20.5+incompatible // indirect
	github.com/shopspring/decimal v1.2.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200824131525-c12d262b63d8 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/filecoin-project/specs-actors => github.com/EpiK-Protocol/go-epik-actors v0.6.2-0.20201022092154-67fcbed36c3a

replace github.com/supranational/blst => github.com/supranational/blst v0.1.2-alpha.1
