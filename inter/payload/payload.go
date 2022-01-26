package payload

import (
	"DictGenerate/inter/pinyin"
	"strings"
	"time"
)

// ToUpperLower 转换大小写
func ToUpperLower(s string) (low, up string) {
	return strings.ToLower(s), strings.ToUpper(s)
}

// ToUpperLowerSlice 转换大小写 - 切片
func ToUpperLowerSlice(s []string) (low, up []string) {
	total := len(s)
	low = make([]string, total)
	up = make([]string, total)
	for i := 0; i < total; i++ {
		low[i], up[i] = ToUpperLower(s[i])
	}
	return
}

// UcWords 字符串首字母大写
func UcWords(s string) string {
	var upperStr string
	vv := []rune(s) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return s
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

// UcWordsSlice 字符串首字母大写 - 切片
func UcWordsSlice(s []string) []string {
	total := len(s)
	list := make([]string, len(s))
	for i := 0; i < total; i++ {
		list[i] = UcWords(s[i])
	}
	return list
}

// SliceUnique 切片去重 - map
func SliceUnique(s []string) []string {
	total := len(s)
	result := make([]string, 0, total)
	tempMap := make(map[string]struct{})
	for i := 0; i < total; i++ {
		if _, is := tempMap[s[i]]; !is {
			tempMap[s[i]] = struct{}{}
			result = append(result, s[i])
		}
	}
	return result
}

// ReverseString 反转字符串
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// FirstString 获取字符串首字母
func FirstString(s string) string {
	return string([]rune(s)[0])
}

// FirstSlice 获取字符串首字母
func FirstSlice(s []string) []string {
	total := len(s)
	mix := make([]string, 0, total)

	for i := 0; i < total; i++ {
		mix = append(mix, FirstString(s[i]))
	}
	return mix
}

// MixName 混淆姓名
func MixName(s []string) []string {
	total := len(s)
	mixList := make([]string, 0, total*2+8)
	lower, upper := ToUpperLowerSlice(s)

	// 组合全名 例：zhou jie lun 组合 zhoujielun
	fullNameLower := strings.Join(lower, "")
	fullNameUpper := strings.Join(upper, "")
	firstUpperStr := UcWords(fullNameLower)
	mixList = append(mixList, fullNameLower, fullNameUpper, firstUpperStr)

	// 组合全名 姓放最后 例：zhou jie lun 组合 jielunzhou
	if total > 1 {
		fullNameLowerX := strings.Join(lower[1:], "") + lower[0]
		fullNameUpperX := strings.Join(upper[1:], "") + upper[0]

		mixList = append(mixList, fullNameLowerX, fullNameUpperX)
	}

	// 组合后两个
	if total >= 2 {
		// 正序组合 例：zhou jie lun 组合 jielun
		givenNameLower := lower[total-2] + lower[total-1]
		givenNameUpper := upper[total-2] + upper[total-1]
		// 倒序组合 例：zhou jie lun 组合 lunjie
		givenNameLowerInverted := lower[total-1] + lower[total-2]
		givenNameUpperInverted := upper[total-1] + upper[total-2]

		mixList = append(mixList, givenNameLower, givenNameUpper, givenNameLowerInverted, givenNameUpperInverted)
	}

	// 组合首字母
	if total > 0 {
		// 首字母切片
		firstLetter := FirstSlice(s)

		// 组合姓和首字母 例：zhou jie lun 组合 zhoujl jlzhou
		nameFirst1Lower, nameFirst1Upper := ToUpperLower(s[0] + strings.Join(firstLetter[1:], ""))
		nameFirst2Lower, nameFirst2Upper := ToUpperLower(strings.Join(firstLetter[1:], "") + s[0])
		mixList = append(mixList, nameFirst1Lower, nameFirst1Upper, nameFirst2Lower, nameFirst2Upper)

		if total > 1 {
			// 组合首字母和姓 例：zhou jie lun 组合 zjielun jielunz
			firstName1Lower, firstName1Upper := ToUpperLower(firstLetter[0] + strings.Join(s[1:], ""))
			firstName2Lower, firstName2Upper := ToUpperLower(strings.Join(s[1:], "") + firstLetter[0])
			// 首字母大写
			firstUpperStr := UcWords(firstName1Lower)
			firstUpperStr2 := UcWords(firstName2Lower)
			mixList = append(mixList, firstName1Lower, firstName1Upper, firstName2Lower, firstName2Upper, firstUpperStr, firstUpperStr2)
		}
	}

	// 合并大小写
	mixList = append(mixList, lower...)
	mixList = append(mixList, upper...)
	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixFirstLetter 混淆姓名首字母
func MixFirstLetter(s string) []string {
	mixList := make([]string, 0, 4)

	// 大小写
	lower, upper := ToUpperLower(s)
	// 首字母大写
	firstUpperStr := UcWords(lower)

	// 反转字符串
	lowerRe := ReverseString(lower)
	upperRe := ReverseString(upper)

	// 重复
	repeat2 := strings.Repeat(s, 2)
	lowerRepeat, upperRepeat := ToUpperLower(repeat2)

	// 去重
	mixList = append(mixList, lower, upper, lowerRe, upperRe, lowerRepeat, upperRepeat, firstUpperStr)
	mixList = SliceUnique(mixList)
	return mixList
}

// MixUsername 混淆用户名
func MixUsername(s string) []string {
	mixList := make([]string, 0)
	lower, upper := ToUpperLowerSlice(strings.Split(s, ","))

	// 首字母大写
	firstUpperStr := UcWordsSlice(lower)

	// 去重
	mixList = append(mixList, lower...)
	mixList = append(mixList, upper...)
	mixList = append(mixList, firstUpperStr...)
	mixList = SliceUnique(mixList)
	return mixList
}

// MixBirthday 混淆生日
func MixBirthday(birthday, lunar string) []string {
	mixList := make([]string, 16)

	birthdayTime, _ := time.Parse("20060102", birthday)
	lunarTime, _ := time.Parse("20060102", lunar)

	mixList[0] = birthdayTime.Format("20060102")
	mixList[1] = birthdayTime.Format("200601")
	mixList[2] = birthdayTime.Format("0102")
	mixList[3] = birthdayTime.Format("2006")
	mixList[4] = birthdayTime.Format("01")
	mixList[5] = birthdayTime.Format("02")
	mixList[6] = lunarTime.Format("20060102")
	mixList[7] = lunarTime.Format("200601")
	mixList[8] = lunarTime.Format("0102")
	mixList[9] = lunarTime.Format("2006")
	mixList[10] = lunarTime.Format("01")
	mixList[11] = lunarTime.Format("02")

	// 19910812 组合 910812
	mixList[12] = string([]rune(mixList[3])[2:]) + mixList[2]
	mixList[13] = string([]rune(mixList[9])[2:]) + mixList[8]
	// 19910812 组合 9108
	mixList[14] = string([]rune(mixList[3])[2:]) + mixList[4]
	mixList[15] = string([]rune(mixList[9])[2:]) + mixList[10]

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixEmail 混淆邮箱地址
func MixEmail(s string) []string {
	mixList := make([]string, 5)

	email := strings.Split(s, "@")
	mixList[0], mixList[1] = ToUpperLower(email[0])
	mixList[2], mixList[3] = ToUpperLower(s)

	// 首字母大写
	mixList[4] = UcWords(mixList[0])

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixMobile 混淆手机号
func MixMobile(s string) []string {
	mixList := make([]string, 6)

	mobile := []rune(s)
	mixList[0] = string(mobile)
	mixList[1] = string(mobile[3:7])
	mixList[2] = string(mobile[7:])
	mixList[3] = string(mobile[5:])
	mixList[4] = string(mobile[6:])
	mixList[5] = string(mobile[8:])

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixIdentityCard 混淆身份证
func MixIdentityCard(s string) []string {
	mixList := make([]string, 19)

	card := []rune(s)
	mixList[0] = string(card[:6])
	mixList[1] = string(card[10:])
	mixList[2] = string(card[12:])
	mixList[3] = string(card[14:])
	mixList[4] = string(card[15:])
	mixList[5] = string(card[8:])
	mixList[6] = string(card[9:17])
	mixList[7] = string(card[11:17])
	mixList[8] = string(card[13:17])
	mixList[9] = string(card[14:17])
	mixList[10] = string(card[8:14])
	mixList[11] = string(card[10:14])
	mixList[12] = string(card[6:12])
	mixList[13] = string(card[6:14])
	mixList[14] = string(card[11:14])
	mixList[15] = string(card[10:12])
	mixList[16] = string(card[6:10])
	mixList[17] = string(card[10:12])
	mixList[18] = string(card[12:14])

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixPhrase 混淆短语
func MixPhrase(s string) []string {
	mixList := make([]string, 0)

	list := strings.Split(s, ",")
	lower, upper := ToUpperLowerSlice(list)

	// 首字母大写
	firstUpperStr := UcWordsSlice(lower)

	// 去重
	mixList = append(mixList, lower...)
	mixList = append(mixList, upper...)
	mixList = append(mixList, firstUpperStr...)
	mixList = SliceUnique(mixList)
	return mixList
}

// MixWordGroup 混淆词组
func MixWordGroup(s string) []string {
	mixList := make([]string, 0)
	list := strings.Split(s, ",")

	lower, upper := ToUpperLowerSlice(list)
	firstUpperStr := UcWordsSlice(lower)
	mixList = append(mixList, lower...)
	mixList = append(mixList, upper...)
	mixList = append(mixList, firstUpperStr...)

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixConnector 连接符
func MixConnector(s string) []string {
	mixList := strings.Split(s, "")

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}

// MixCompany 混淆公司/组织名称
func MixCompany(s []string) []string {
	mixList := make([]string, 0)

	lower, upper := ToUpperLowerSlice(s)
	mixList = append(mixList, lower...)
	mixList = append(mixList, upper...)

	// 公司/组织 首字母
	companyFirstLetter := pinyin.FormatSliceFirstLetter(s)
	low, upp := ToUpperLower(companyFirstLetter)
	mixList = append(mixList, low, upp)

	// 去重
	mixList = SliceUnique(mixList)
	return mixList
}
