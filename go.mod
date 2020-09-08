module github.com/EpiK-Protocol/epik-explorer-backend

go 1.14

require (
	github.com/EpiK-Protocol/go-epik v0.4.2-0.20200901164337-9b91560725db
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ethereum/go-ethereum v1.9.20
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/filecoin-project/go-address v0.0.2-0.20200504173055-8b6f2fb2b3ef
	github.com/filecoin-project/go-jsonrpc v0.1.1-0.20200602181149-522144ab4e24
	github.com/filecoin-project/specs-actors v0.6.2-0.20200724193152-534b25bdca30
	github.com/gin-gonic/gin v1.6.3
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/ipfs/go-cid v0.0.6
	github.com/jackc/pgproto3/v2 v2.0.4 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.3.0+incompatible
	github.com/lestrrat-go/strftime v1.0.3 // indirect
	github.com/multiformats/go-multibase v0.0.3
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/tebeka/strftime v0.1.5 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	gopkg.in/yaml.v2 v2.3.0
	gorm.io/driver/postgres v1.0.0
	gorm.io/gorm v1.20.0
)

replace github.com/filecoin-project/specs-actors => ../go-epik-actors

replace github.com/ethereum/go-ethereum => ./extern/go-ethereum

replace github.com/supranational/blst => github.com/supranational/blst v0.1.2-alpha.1
