package config

// 版本信息
var (
	// IsDebug = "0"
	App       = "jNacos"
	IsDebug   = "1"                        // 调试模式
	Version   = "0.1.0"                    // 版本号
	GitTag    = "2022.02.27.dev"           // 代码版本
	BuildTime = "2022-02-27T23:09:57+0800" // 编译时间
)

const (
	SCHEME_SYS_CONFIG = "SYSCONF://" // 配置依赖加载
	SCHEME_SYS_MEMORY = "SYSMEM://"  // 输出到内存，以环境变量形式提供
)

type Config struct {
	NacosCfg  NacosConfig      `json:"nacos" toml:"nacos" mapstructure:"nacos"`
	NacosJobs []NacosJobConfig `json:"nacosJobs" toml:"nacosJobs" mapstructure:"nacosJobs"`
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

	// OpenKMS              bool // it's to open kms,default is false. https://help.aliyun.com/product/28933.html
	// NotLoadCacheAtStart  bool // not to load persistent nacos service info in CacheDir at start time
	// UpdateCacheWhenEmpty bool // update cache when get empty service instance from server

	Username    string `json:"username" toml:"username" mapstructure:"username"`          // the username for nacos auth
	Password    string `json:"password" toml:"password" mapstructure:"password"`          // the password for nacos auth
	ContextPath string `json:"contextPath" toml:"contextPath" mapstructure:"contextPath"` // the nacos server contextpath

	LogLevel string `json:"logLevel" toml:"logLevel" mapstructure:"logLevel"` //
}

func (nc *NacosConfig) Equals(newConf *NacosConfig) bool {
	return nc.AccessKey == newConf.AccessKey &&
		nc.Endpoint == newConf.Endpoint &&
		nc.NamespaceId == newConf.NamespaceId &&
		nc.NamespaceId == newConf.SecretKey &&
		nc.AppName == newConf.AppName &&
		nc.RegionId == newConf.RegionId &&
		nc.Username == newConf.Username &&
		nc.Password == newConf.Password &&
		nc.ContextPath == newConf.ContextPath
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
	// ExecOnStart bool
}

// TODO 判断两个对象是否相等
func (nc *NacosJobConfig) Equals(newConf *NacosJobConfig) bool {
	if nc.Exec == newConf.Exec {
		// 需要判断数据是否相等

	}
	return false
}

type NacosJobFileConfig struct {
	DataId string `json:"dataId" toml:"dataId" mapstructure:"dataId"`
	Group  string `json:"group" toml:"group" mapstructure:"group"`
	// 输出文件
	// 特殊文件：SYSCONF://{可选配置Id} 用于重新加载配置
	Outfile string `json:"outfile" toml:"outfile" mapstructure:"outfile"`
}

// 直接访问资源（用于 fetch、push）
type DirectConfig struct {
	//config.NacosConfig `mapstructure:",squash"`
	NacosCfg *NacosConfig `json:"nacos" toml:"nacos" mapstructure:"nacos"` // 连接配置

	Group  string `json:"group" toml:"group" mapstructure:"group"`    // 资源组
	DataId string `json:"dataId" toml:"dataId" mapstructure:"dataId"` // 资源 ID
	File   string `json:"file" toml:"file" mapstructure:"file"`       // 输入或输出文件
}
