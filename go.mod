module rtoken-swap

go 1.16

require (
	github.com/ChainSafe/log15 v1.0.0
	github.com/JFJun/go-substrate-crypto v1.0.1
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/ethereum/go-ethereum v1.10.6
	github.com/gorilla/websocket v1.4.2
	github.com/huandu/xstrings v1.3.2
	github.com/itering/scale.go v1.1.39
	github.com/itering/substrate-api-rpc v0.3.5
	github.com/mr-tron/base58 v1.2.0
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stafiprotocol/chainbridge v1.0.0
	github.com/stafiprotocol/go-substrate-rpc-client v1.1.3
	github.com/stafiprotocol/solana-go-sdk v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.9
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871
	golang.org/x/sys v0.0.0-20211124211545-fe61309f8881 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
