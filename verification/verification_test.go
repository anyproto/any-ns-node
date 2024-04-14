package verification

import (
	"testing"

	"github.com/zeebo/assert"
)

func TestAAS_VerifyAdminIdentity(t *testing.T) {
	t.Run("fail if wrong key", func(t *testing.T) {
		realPeerKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="
		adminIdentity := "psqF9Jj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="

		err := VerifyAdminIdentity(realPeerKey, adminIdentity)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// 2 - sign it
		realPeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="
		adminIdentity := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := VerifyAdminIdentity(realPeerKey, adminIdentity)
		assert.NoError(t, err)
	})
}
