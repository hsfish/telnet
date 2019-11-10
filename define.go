package telnet

const (
	COMMAND_SUCCESS = 0
	COMMAND_FAILED  = 1
	COMMAND_MORE    = 2

	DIAL_TIMEOUT_DEFAULT   = 5
	TIMEOUT_DEFAULT        = 10
	USER_REGEX_DEFAULT     = "name|login"
	PASSWORD_REGEX_DEFAULT = "ssword"
	REGEX_DEFAULT          = "\\S(#|>|]|:)$"
)

type ClientConf struct {
	Host          string
	Port          int
	User          string
	Password      string
	Timeout       int // s
	DialTimeout   int // s
	UserRegex     string
	PasswordRegex string
	Pattern       string
}

type CommandBack struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Command struct {
	Cmd     string `json:"cmd"`
	Pattern string `json:"pattern"`
}

type Options struct {
	Timeout   int64  // 执行单条命令返回超时时间 s
	Regex     string // 匹配规则返回
	TrimFirst bool   // 去掉第一条
}
