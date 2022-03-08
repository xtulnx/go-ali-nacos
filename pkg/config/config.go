package config

// 版本信息
var (
	//IsDebug = "0"
	IsDebug   = "1"                        // 调试模式
	Version   = "0.1.0"                    // 版本号
	GitTag    = "2022.02.27.dev"           // 代码版本
	BuildTime = "2022-02-27T23:09:57+0800" // 编译时间
)

type Config struct {
	NacosCfg  NacosConfig      `json:"nacos" toml:"nacos" mapstructure:"nacos"`
	NacosJobs []NacosJobConfig `json:"nacosJobs" toml:"nacosJobs" mapstructure:"nacosJobs"`

	// 自动刷新周期，单位秒
	AutoRefresh int
}

type NacosConfig struct {
	// 从控制台命名空间管理的"命名空间详情"中拷贝 End Point、命名空间 ID
	Endpoint    string `json:"endpoint" toml:"endpoint" mapstructure:"endpoint"`
	NamespaceId string `json:"namespaceId" toml:"namespaceId" mapstructure:"namespaceId"`

	// 推荐使用 RAM 用户的 accessKey、secretKey
	AccessKey string `json:"accessKey" toml:"accessKey" mapstructure:"accessKey"`
	SecretKey string `json:"secretKey" toml:"secretKey" mapstructure:"secretKey"`

	// 可选

	AppName  string `json:"appName" toml:"appName" mapstructure:"appName"`    // the appName
	RegionId string `json:"regionId" toml:"regionId" mapstructure:"regionId"` // the regionId for kms

	//OpenKMS              bool // it's to open kms,default is false. https://help.aliyun.com/product/28933.html
	//NotLoadCacheAtStart  bool // not to load persistent nacos service info in CacheDir at start time
	//UpdateCacheWhenEmpty bool // update cache when get empty service instance from server

	Username    string `json:"username" toml:"username" mapstructure:"username"`          // the username for nacos auth
	Password    string `json:"password" toml:"password" mapstructure:"password"`          // the password for nacos auth
	ContextPath string `json:"contextPath" toml:"contextPath" mapstructure:"contextPath"` // the nacos server contextpath
}

// 任务子节点
type NacosJobConfig struct {
	// ~~
	File []NacosJobFileConfig `json:"file" toml:"file" mapstructure:"file"`
	// 执行外部命令
	Exec   string   `json:"exec" toml:"exec" mapstructure:"exec"`
	Params []string `json:"params" toml:"params" mapstructure:"params"`
	// 任务执行超时时间，默认 30s
	Timeout int `json:"timeout" toml:"timeout" mapstructure:"timeout"`
	// 启动时运行
	//ExecOnStart bool
}

type NacosJobFileConfig struct {
	DataId string `json:"dataId" toml:"dataId" mapstructure:"dataId"`
	Group  string `json:"group" toml:"group" mapstructure:"group"`
	// 输出文件
	Outfile string `json:"outfile" toml:"outfile" mapstructure:"outfile"`
}
