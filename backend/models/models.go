package models

// 请求和响应模型定义

// SheetMetaRequest 获取表结构请求
type SheetMetaRequest struct {
	RequestID string  `json:"requestId"`
	Params    string  `json:"params"` // JSON字符串
	Context   Context `json:"context"`
}

// RecordsRequest 获取表记录请求
type RecordsRequest struct {
	RequestID  string  `json:"requestId"`
	MaxResults int     `json:"maxResults"`
	NextToken  string  `json:"nextToken"`
	Params     string  `json:"params"` // JSON字符串
	Context    Context `json:"context"`
}

// Context 请求上下文
type Context struct {
	UnionID string `json:"unionId"`
	CorpID  string `json:"corpId"`
}

// MySQLConfig MySQL连接配置(从params中解析)
type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Table    string `json:"table"`
}

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// SheetMetaResponse 表结构响应数据
type SheetMetaResponse struct {
	SheetName string  `json:"sheetName"`
	Fields    []Field `json:"fields"`
}

// Field 字段定义
type Field struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	IsPrimary   bool                   `json:"isPrimary"`
	Property    map[string]interface{} `json:"property,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// RecordsResponse 表记录响应数据
type RecordsResponse struct {
	NextToken string   `json:"nextToken,omitempty"`
	HasMore   bool     `json:"hasMore"`
	Records   []Record `json:"records"`
	Total     int      `json:"total,omitempty"`
}

// Record 单条记录
type Record struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// 错误码定义
const (
	CodeSuccess          = 0
	CodeParamError       = 10001 // 参数错误
	CodeConfigError      = 10002 // 配置信息错误
	CodeInsufficientAuth = 10003 // 权益不足
	CodeAuthFailed       = 10004 // 身份校验失败
	CodeThirdPartyError  = 10005 // 三方系统异常
)
