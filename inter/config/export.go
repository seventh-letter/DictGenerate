package config

type configJSONExport struct {
	Name 				[]string	`json:"name"`					// 姓名(全拼)
	FirstLetter 		string 		`json:"first_letter"`			// 姓名首字母
	Short 				[]string 	`json:"short"`					// 短名
	Username			[]string 	`json:"username"`				// 常用用户名
	Birthday 			string 		`json:"birthday"`				// 公历生日 yyyymmdd
	Lunar 				string 		`json:"lunar"`					// 农历生日 yyyymmdd
	Email				string 		`json:"email"`					// 邮箱
	Mobile				[]string 	`json:"mobile"`					// 手机号
	QQ					[]string 	`json:"qq"`						// QQ
	Company				[]string 	`json:"company"`				// 公司(英文)
	Phrase				string 		`json:"phrase"`					// 短语（英文）
	IdentityCard		string 		`json:"identity_card"`			// 身份证 18位
	JobNumber			string 		`json:"job_number"`				// 工号
	WordGroup			string 		`json:"word_group"`				// 常用词组
	Connector			string 		`json:"connector"`				// 连接符 @#.-_~!?%&*+=$/|
	FilterNumber		bool   		`json:"filter_number"`			// 过滤纯数字
	FilterLetter		bool   		`json:"filter_letter"`			// 过滤纯字母
	FilterLenMin		int    		`json:"filter_len_min"`			// 过滤长度 - 最小值
	FilterLenMax		int    		`json:"filter_len_max"`			// 过滤长度 - 最大值
}

func NewConfigJSONExport() *configJSONExport {
	return &configJSONExport{
		Phrase: Phrase,
		WordGroup: WordGroup,
		Connector: Connector,
		FilterNumber: false,
		FilterLetter: false,
		FilterLenMin: 6,
		FilterLenMax: 12,
	}
}