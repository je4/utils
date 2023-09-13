package elastic8

import configutil "github.com/je4/utils/v2/pkg/config"

type Elastic8Config struct {
	Adresses               []string
	Username               configutil.EnvString
	Password               configutil.EnvString
	CACert                 string
	CertificateFingerprint configutil.EnvString // 2D:51:25:AB:34:0B:57:83:BC:79:3B:A9:2D:B2:6E:63:95:FC:BE:ED:6A:2E:00:F8:11:FE:75:B6:9A:EF:2B:0A
	ServiceToken           configutil.EnvString
	APIKey                 configutil.EnvString
	CloudID                string
}
