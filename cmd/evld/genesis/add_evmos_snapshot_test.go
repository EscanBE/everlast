package genesis

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"
)

func Test_getAirdropAmountFromEvmosSnapshotBalance(t *testing.T) {
	const magicNumber float64 = 2
	const keepRatio float64 = 0.5
	tests := []struct {
		amount int64
		want   int64
	}{
		{
			amount: 800,
			want:   200,
		},
		{
			amount: 2000,
			want:   500,
		},
		{
			amount: 2040,
			want:   509,
		},
		{
			amount: 3000,
			want:   725,
		},
		{
			amount: 4000,
			want:   925,
		},
		{
			amount: 4010,
			want:   926,
		},
		{
			amount: 5000,
			want:   1100,
		},
		{
			amount: 5010,
			want:   1101,
		},
		{
			amount: 6000,
			want:   1250,
		},
		{
			amount: 6010,
			want:   1251,
		},
		{
			amount: 7000,
			want:   1375,
		},
		{
			amount: 7010,
			want:   1376,
		},
		{
			amount: 9000,
			want:   1575,
		},
		{
			amount: 9010,
			want:   1575,
		},
		{
			amount: 13000,
			want:   1875,
		},
		{
			amount: 13010,
			want:   1875,
		},
		{
			amount: 19000,
			want:   2175,
		},
		{
			amount: 19010,
			want:   2175,
		},
		{
			amount: 39000,
			want:   2675,
		},
		{
			amount: 39020,
			want:   2675,
		},
		{
			amount: 239000,
			want:   5175,
		},
		{
			amount: 239100,
			want:   5175,
		},
		{
			amount: 339000,
			want:   5425,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d => %d", tt.amount, tt.want), func(t *testing.T) {
			got := getEffectiveAmountFromEvmosSnapshotBalance(sdkmath.NewInt(tt.amount).MulRaw(1e18), magicNumber, keepRatio)
			require.Equal(t, tt.want, got.QuoRaw(1e18).Int64())
		})
	}
}
