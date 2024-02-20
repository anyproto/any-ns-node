package verification

import (
	"testing"

	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

func TestAAS_VerifyAdminIdentity(t *testing.T) {
	t.Run("fail", func(t *testing.T) {

		// 0 - garbage data test
		realSignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		err := VerifyAdminIdentity(realSignKey, []byte("payload"), []byte("signature"))
		assert.Error(t, err)

		// 1 - pack some structure
		nrr := nsp.AdminFundUserAccountRequest{
			OwnerEthAddress: "",
			NamesCount:      0,
		}

		marshalled, err := nrr.Marshal()
		require.NoError(t, err)

		// 2 - sign it with some random (wrong) key
		accountKeys, err := accountdata.NewRandom()
		require.NoError(t, err)

		sig, err := accountKeys.SignKey.Sign(marshalled)
		require.NoError(t, err)

		err = VerifyAdminIdentity(realSignKey, marshalled, sig)
		assert.Error(t, err)
	})

	t.Run("fail if wrong key", func(t *testing.T) {
		// 1 - pack some structure
		nrr := nsp.AdminFundUserAccountRequest{
			OwnerEthAddress: "",
			NamesCount:      0,
		}

		marshalled, err := nrr.Marshal()
		require.NoError(t, err)

		// 2 - sign it
		badKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="

		signKey, err := crypto.DecodeKeyFromString(
			// see here:
			badKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		require.NoError(t, err)

		sig, err := signKey.Sign(marshalled)
		require.NoError(t, err)

		realSignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="
		err = VerifyAdminIdentity(realSignKey, marshalled, sig)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// 1 - pack some structure
		nrr := nsp.AdminFundUserAccountRequest{
			OwnerEthAddress: "",
			NamesCount:      0,
		}

		marshalled, err := nrr.Marshal()
		require.NoError(t, err)

		// 2 - sign it
		realSignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		signKey, err := crypto.DecodeKeyFromString(
			realSignKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		require.NoError(t, err)

		sig, err := signKey.Sign(marshalled)
		require.NoError(t, err)

		// get associated pub key
		//pubKey := signKey.GetPublic()
		// identity str
		//identityStr := pubKey.Account()
		// A5ommzwhpR5ngp11q9q1P2MMzhUE46Hi421RJbPqswALyoyr
		//log.Info("identity", zap.String("identity", identityStr))

		err = VerifyAdminIdentity(realSignKey, marshalled, sig)
		assert.NoError(t, err)
	})
}
