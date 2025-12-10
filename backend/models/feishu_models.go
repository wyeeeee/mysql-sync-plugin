package models

// 飞书多维表格数据源接口模型定义

// 飞书字段类型枚举
const (
	FeishuFieldTypeText       = 1  // 多行文本
	FeishuFieldTypeNumber     = 2  // 数字
	FeishuFieldTypeSingleSel  = 3  // 单选
	FeishuFieldTypeMultiSel   = 4  // 多选
	FeishuFieldTypeDate       = 5  // 日期
	FeishuFieldTypeBarcode    = 6  // 条码
	FeishuFieldTypeCheckbox   = 7  // 复选框
	FeishuFieldTypeCurrency   = 8  // 货币
	FeishuFieldTypePhone      = 9  // 电话号码
	FeishuFieldTypeHyperlink  = 10 // 超链接
	FeishuFieldTypeProgress   = 11 // 进度
	FeishuFieldTypeRating     = 12 // 评分
	FeishuFieldTypeLocation   = 13 // 地理位置
)

// 飞书错误码
const (
	FeishuCodeSuccess       = 0       // 成功
	FeishuCodeConfigError   = 1254400 // 配置错误
	FeishuCodeAuthError     = 1254403 // 权限异常
	FeishuCodeSystemError   = 1254500 // 系统异常
	FeishuCodeRateLimitError = 1254501 // 限流异常
	FeishuCodePaymentError  = 1254505 // 付费错误
)

// FeishuTableMetaRequest 飞书获取表结构请求
type FeishuTableMetaRequest struct {
	Params  string `json:"params"`  // JSON字符串，包含 datasourceConfig
	Context string `json:"context"` // JSON字符串，系统参数（需要二次解析）
}

// FeishuRecordsRequest 飞书获取表记录请求
type FeishuRecordsRequest struct {
	Params  string `json:"params"`  // JSON字符串，包含分页参数
	Context string `json:"context"` // JSON字符串，系统参数（需要二次解析）
}

// FeishuRequestContext 飞书请求上下文
type FeishuRequestContext struct {
	Bitable       FeishuBitableContext `json:"bitable"`
	PackID        string               `json:"packID"`
	Type          string               `json:"type"`
	TenantKey     string               `json:"tenantKey"`
	UserTenantKey string               `json:"userTenantKey"`
	BizInstanceID string               `json:"bizInstanceID"`
	ScriptArgs    FeishuScriptArgs     `json:"scriptArgs"`
}

// FeishuBitableContext 飞书Bitable上下文
type FeishuBitableContext struct {
	Token string `json:"token"`
	LogID string `json:"logID"`
}

// FeishuScriptArgs 飞书脚本参数
type FeishuScriptArgs struct {
	ProjectURL string `json:"projectURL"`
	BaseOpenID string `json:"baseOpenID"` // 用户身份标识
}

// FeishuParams 飞书请求参数（从params字段解析）
type FeishuParams struct {
	DatasourceConfig string `json:"datasourceConfig"` // 数据源配置JSON字符串
	TransactionID    string `json:"transactionID"`    // 同步事务ID
	PageToken        string `json:"pageToken"`        // 分页token
	MaxPageSize      int    `json:"maxPageSize"`      // 最大页大小
}

// FeishuResponse 飞书通用响应结构
type FeishuResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// FeishuErrorMsg 飞书错误消息（需要中英文）
type FeishuErrorMsg struct {
	Zh string `json:"zh"`
	En string `json:"en"`
}

// FeishuTableMetaResponse 飞书表结构响应数据
type FeishuTableMetaResponse struct {
	TableName string        `json:"tableName"`
	Fields    []FeishuField `json:"fields"`
}

// FeishuField 飞书字段定义
type FeishuField struct {
	FieldID     string                 `json:"fieldID"`
	FieldName   string                 `json:"fieldName"`
	FieldType   int                    `json:"fieldType"`
	IsPrimary   bool                   `json:"isPrimary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Property    map[string]interface{} `json:"property,omitempty"`
}

// FeishuRecordsResponse 飞书表记录响应数据
type FeishuRecordsResponse struct {
	NextPageToken string         `json:"nextPageToken,omitempty"`
	HasMore       bool           `json:"hasMore"`
	Records       []FeishuRecord `json:"records"`
}

// FeishuRecord 飞书单条记录
type FeishuRecord struct {
	PrimaryID string                 `json:"primaryID"`
	Data      map[string]interface{} `json:"data"`
}

// DingtalkFieldTypeToFeishu 钉钉字段类型转飞书字段类型
func DingtalkFieldTypeToFeishu(dingtalkType string) int {
	switch dingtalkType {
	case "text":
		return FeishuFieldTypeText
	case "number":
		return FeishuFieldTypeNumber
	case "singleSelect":
		return FeishuFieldTypeSingleSel
	case "multiSelect":
		return FeishuFieldTypeMultiSel
	case "dateTime":
		return FeishuFieldTypeDate
	case "barcode":
		return FeishuFieldTypeBarcode
	case "checkbox":
		return FeishuFieldTypeCheckbox
	case "currency":
		return FeishuFieldTypeCurrency
	case "phone":
		return FeishuFieldTypePhone
	case "hyperlink":
		return FeishuFieldTypeHyperlink
	case "progress":
		return FeishuFieldTypeProgress
	case "rating":
		return FeishuFieldTypeRating
	case "location":
		return FeishuFieldTypeLocation
	default:
		return FeishuFieldTypeText // 默认为文本类型
	}
}

// DingtalkErrorCodeToFeishu 钉钉错误码转飞书错误码
func DingtalkErrorCodeToFeishu(dingtalkCode int) int {
	switch dingtalkCode {
	case CodeSuccess:
		return FeishuCodeSuccess
	case CodeParamError, CodeConfigError:
		return FeishuCodeConfigError
	case CodeInsufficientAuth, CodeAuthFailed:
		return FeishuCodeAuthError
	case CodeThirdPartyError:
		return FeishuCodeSystemError
	default:
		return FeishuCodeSystemError
	}
}

// ConvertToFeishuTableMeta 将钉钉表结构响应转换为飞书格式
func ConvertToFeishuTableMeta(dingtalk *SheetMetaResponse) *FeishuTableMetaResponse {
	return ConvertToFeishuTableMetaWithMappings(dingtalk, nil)
}

// ConvertToFeishuTableMetaWithMappings 将钉钉表结构响应转换为飞书格式，并应用字段映射
func ConvertToFeishuTableMetaWithMappings(dingtalk *SheetMetaResponse, mappings []FieldMapping) *FeishuTableMetaResponse {
	// 构建别名映射表
	aliasMap := make(map[string]string)
	for _, m := range mappings {
		if m.AliasField != "" && m.AliasField != m.MysqlField {
			aliasMap[m.MysqlField] = m.AliasField
		}
	}

	feishuFields := make([]FeishuField, len(dingtalk.Fields))
	primaryFound := false // 飞书只允许一个主键字段

	for i, f := range dingtalk.Fields {
		// 飞书要求 fieldID 只能包含英文、数字、下划线
		fieldID := sanitizeFieldID(f.ID)

		// fieldName 使用别名（如果有的话）
		fieldName := f.Name
		if alias, ok := aliasMap[f.Name]; ok {
			fieldName = alias
		}

		// 转换 property 为飞书格式
		feishuProperty := convertPropertyToFeishu(f.Property, DingtalkFieldTypeToFeishu(f.Type))

		// 飞书只允许一个主键字段，取第一个主键
		isPrimary := false
		if f.IsPrimary && !primaryFound {
			isPrimary = true
			primaryFound = true
		}

		feishuFields[i] = FeishuField{
			FieldID:     fieldID,
			FieldName:   fieldName,
			FieldType:   DingtalkFieldTypeToFeishu(f.Type),
			IsPrimary:   isPrimary,
			Description: f.Description,
			Property:    feishuProperty,
		}
	}

	// 飞书要求必须有一个主键字段，如果没有找到主键，将第一个字段设为主键
	if !primaryFound && len(feishuFields) > 0 {
		feishuFields[0].IsPrimary = true
	}

	return &FeishuTableMetaResponse{
		TableName: dingtalk.SheetName,
		Fields:    feishuFields,
	}
}

// sanitizeFieldID 清理字段ID，确保只包含英文、数字、下划线
func sanitizeFieldID(id string) string {
	var result []rune
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return "field_0"
	}
	return string(result)
}

// convertPropertyToFeishu 转换属性为飞书格式
func convertPropertyToFeishu(property map[string]interface{}, fieldType int) map[string]interface{} {
	if property == nil {
		return nil
	}

	result := make(map[string]interface{})

	// 根据字段类型转换 formatter
	switch fieldType {
	case FeishuFieldTypeNumber:
		// 飞书数字格式: "0", "0.0", "0.00", "#,##0" 等
		if formatter, ok := property["formatter"].(string); ok {
			result["formatter"] = convertNumberFormatter(formatter)
		} else {
			result["formatter"] = "0"
		}
	case FeishuFieldTypeDate:
		// 飞书日期格式
		if formatter, ok := property["formatter"].(string); ok {
			result["formatter"] = formatter
		} else {
			result["formatter"] = "yyyy/MM/dd"
		}
	case FeishuFieldTypeCurrency:
		if formatter, ok := property["formatter"].(string); ok {
			result["formatter"] = formatter
		}
		if code, ok := property["currencyCode"].(string); ok {
			result["currencyCode"] = code
		} else {
			result["currencyCode"] = "CNY"
		}
	default:
		// 其他类型直接复制
		for k, v := range property {
			result[k] = v
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// convertNumberFormatter 转换数字格式化器
func convertNumberFormatter(formatter string) string {
	// 钉钉格式到飞书格式的映射
	switch formatter {
	case "INT", "int":
		return "0"
	case "FLOAT", "float":
		return "0.00"
	case "DECIMAL", "decimal":
		return "0.00"
	default:
		// 如果已经是飞书格式，直接返回
		if formatter == "0" || formatter == "0.0" || formatter == "0.00" ||
			formatter == "0.000" || formatter == "0.0000" ||
			formatter == "#,##0" || formatter == "#,##0.00" ||
			formatter == "0%" || formatter == "0.00%" {
			return formatter
		}
		return "0"
	}
}

// ConvertToFeishuRecords 将钉钉记录响应转换为飞书格式
func ConvertToFeishuRecords(dingtalk *RecordsResponse) *FeishuRecordsResponse {
	feishuRecords := make([]FeishuRecord, len(dingtalk.Records))
	for i, r := range dingtalk.Records {
		// 清理记录中的字段key，确保与表结构中的fieldID一致
		sanitizedData := make(map[string]interface{})
		for key, value := range r.Fields {
			sanitizedKey := sanitizeFieldID(key)
			sanitizedData[sanitizedKey] = value
		}

		feishuRecords[i] = FeishuRecord{
			PrimaryID: sanitizeFieldID(r.ID), // primaryID也需要清理
			Data:      sanitizedData,
		}
	}
	return &FeishuRecordsResponse{
		NextPageToken: dingtalk.NextToken,
		HasMore:       dingtalk.HasMore,
		Records:       feishuRecords,
	}
}

// NewFeishuErrorMsg 创建飞书错误消息
func NewFeishuErrorMsg(zh, en string) string {
	// 返回JSON格式的错误消息
	return `{"zh":"` + zh + `","en":"` + en + `"}`
}
