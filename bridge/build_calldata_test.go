package bridge

import (
	"math/big"
	"testing"
)

func TestBuildCallDataUniswap(t *testing.T) {
	type args struct {
		quoteData                     []byte
		tokenOutAddress               string
		srcQty                        *big.Int
		expectedOut                   *big.Int
		proxyContractAddress          string
		incognitoVaultContractAddress string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "sample",
			args: args{
				quoteData:                     []byte(`{"message":"ok","data":{"amountIn":"1","amountOut":"1581.644702","amountOutRaw":"1581644702","blockNumber":"16531363","estimatedGasUsed":"113000","gasPriceWei":"15250969398","gasAdjustedQuoteIn":"1578.917592","gasUsedQuoteToken":"2.727109","gasUsedUSD":"2.725550","route":[[{"type":"V3-pool","poolAddress":"0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640","percent":100,"rawQuote":"1581644702","fee":500,"liquidity":"34772698205548429484","sqrtRatioX96":"1991664897509656491664379768181073","tickCurrent":202653,"tokenIn":{"chainId":1,"decimals":18,"symbol":"WETH","name":"Wrapped Ether","isNative":false,"isToken":true,"address":"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"},"tokenOut":{"chainId":1,"decimals":6,"symbol":"USDC","name":"USD//C","isNative":false,"isToken":true,"address":"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},"amountIn":"1.000000000000000000","amountOut":"1581.644702"}]],"multiRouter":false,"fees":[500],"paths":["0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2","0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"],"percents":[100],"routerString":"[V3] 100.00% = WETH -- 0.05% --> USDC"}}`),
				tokenOutAddress:               "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
				srcQty:                        big.NewInt(1000000000000000000),
				expectedOut:                   big.NewInt(1581644702),
				proxyContractAddress:          "0xe38e54B2d6B1FCdfaAe8B674bF36ca62429fdBDe",
				incognitoVaultContractAddress: "0x43D037A562099A4C2c95b1E2120cc43054450629",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calldata, err := BuildCallDataUniswap(tt.args.quoteData, tt.args.tokenOutAddress, tt.args.srcQty, tt.args.expectedOut, tt.args.proxyContractAddress, tt.args.incognitoVaultContractAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCallDataUniswap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if got != tt.want {
			// 	t.Errorf("BuildCallDataUniswap() = %v, want %v", got, tt.want)
			// }
			t.Logf("Calldata: %v\n", calldata)
		})
	}
}
