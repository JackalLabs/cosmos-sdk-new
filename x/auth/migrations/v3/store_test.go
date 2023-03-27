package v3_test

import (
	"math/rand"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	v1 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v1"
	v4 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v4"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestMigrateMapAccAddressToAccNumberKey test cases for state migration of map to accAddr to accNum
func TestMigrateMapAccAddressToAccNumberKey(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(v1.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	storeService := runtime.NewKVStoreService(storeKey)

	var accountKeeper keeper.AccountKeeper

	app, err := simtestutil.Setup(
		authtestutil.AppConfig,
		&accountKeeper,
	)
	require.NoError(t, err)

	require.NoError(t, v4.Migrate(ctx, storeService, nil, cdc))

	// new base account
	senderPrivKey := secp256k1.GenPrivKey()
	randAccNumber := uint64(rand.Intn(100000-10000) + 10000)
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), randAccNumber, 0)

	ctx = app.BaseApp.NewContext(false, cmtproto.Header{Time: time.Now()})

	// migrator
	m := keeper.NewMigrator(accountKeeper, app.GRPCQueryRouter())
	// set the account to store with map acc addr to acc number
	require.NoError(t, m.V45_SetAccount(ctx, acc))

	testCases := []struct {
		name        string
		doMigration bool
		accNum      uint64
	}{
		{
			name:        "without state migration",
			doMigration: false,
			accNum:      acc.AccountNumber,
		},
		{
			name:        "with state migration",
			doMigration: true,
			accNum:      acc.AccountNumber,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMigration {
				require.NoError(t, m.Migrate2to3(ctx))
			}

			//  get the account address by acc id
			accAddr := accountKeeper.GetAccountAddressByID(ctx, tc.accNum)

			if tc.doMigration {
				require.Equal(t, accAddr, acc.Address)
			} else {
				require.Equal(t, len(accAddr), 0)
			}
		})
	}
}
