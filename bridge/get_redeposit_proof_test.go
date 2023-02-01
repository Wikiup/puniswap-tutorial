package bridge

import (
	"testing"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

func Test_GetProof(t *testing.T) {
	type args struct {
		txhash   string
		endpoint string
	}
	tests := []struct {
		name    string
		args    args
		want    *wcommon.EVMProofRecordData
		want1   *incclient.EVMDepositProof
		wantErr bool
	}{
		{
			name: "sample",
			args: args{
				txhash:   "0x7db0eb153a9571bfd2ca75c42759a89bbbfd7fdd66f7776ff0f980cb661c9774",
				endpoint: "https://eth-fullnode.incognito.org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, proof, err := GetProof(tt.args.txhash, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("getProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("getProof() got = %v, want %v", got, tt.want)
			// }
			// if !reflect.DeepEqual(got1, tt.want1) {
			// 	t.Errorf("getProof() got1 = %v, want %v", got1, tt.want1)
			// }
			t.Logf("Proof: %v\n", proof)
		})
	}
}
