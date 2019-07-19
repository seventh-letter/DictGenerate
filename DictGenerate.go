package main

import (
	"DictGenerate/inter/config"
	"DictGenerate/inter/dictliner"
	"DictGenerate/inter/dictliner/args"
	"DictGenerate/inter/payload"
	"DictGenerate/inter/pinyin"
	"DictGenerate/inter/table"
	"DictGenerate/util"
	"github.com/ernestosuarez/itertools"
	"github.com/nosixtools/solarlunar"
	"github.com/peterh/liner"
	"github.com/telanflow/go-logging"
	"github.com/urfave/cli"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	gLogs = logging.NewLogger(config.AppName)

	reloadFn = func(c *cli.Context) error {
		err := config.C.Reload()
		if err != nil {
			gLogs.Warnf("重载配置错误: %s", err)
		}
		return nil
	}
	saveFunc = func(c *cli.Context) error {
		err := config.C.Save()
		if err != nil {
			gLogs.Warnf("保存配置错误: %s", err)
		}
		return nil
	}
)

func init() {
	util.ChWorkDir()

	// 初始化配置文件
	err := config.C.Init()
	switch err {
	case nil:
	case config.ErrConfigFileNoPermission:
		gLogs.Panicf("配置文件无权限访问：%s", config.ConfigFilePath)
		os.Exit(1)
	case config.ErrConfigContentsParseError:
		gLogs.Panicf("解析Config数据错误：%s", config.ConfigFilePath)
		os.Exit(1)
	default:
		gLogs.Warnf("WARNING: config init error: %s", err)
	}
}

func main() {
	defer config.C.Close()

	app := cli.NewApp()
	app.Name 	= config.AppName
	app.Author 	= config.Author
	app.Version = config.Version
	app.Email	= config.Email
	app.Usage 	= app.Name + " for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Copyright = "(c) 2019 " + app.Author + "."
	app.Description = config.NameChar + `
使用Go语言编写的社工字典生成器`

	app.Action = func(c *cli.Context) {
		if c.NArg() != 0 {
			gLogs.Printf("未找到命令: %s", c.Args().Get(0))
			gLogs.Printf("运行命令 %s help 获取帮助", app.Name)
			return
		}

		var (
			line = dictliner.NewLiner()
			err  error
		)

		line.History, err = dictliner.NewLineHistory(config.HistoryFilePath)
		if err != nil {
			gLogs.Warnf("警告: 读取历史命令文件错误, %s", err)
		}

		line.ReadHistory()
		defer func() {
			line.DoWriteHistory()
			line.Close()
		}()

		// tab 自动补全命令
		line.State.SetCompleter(func(line string) (s []string) {
			var (
				lineArgs = args.Parse(line)
				numArgs  = len(lineArgs)
			)

			for _, cmd := range app.Commands {
				for _, name := range cmd.Names() {
					if !strings.HasPrefix(name, line) {
						continue
					}
					s = append(s, name+" ")
				}
			}

			if numArgs <= 1 {
				return
			}

			s = make([]string, 0)
			thisCmd := app.Command(lineArgs[numArgs-2])
			if thisCmd == nil {
				return
			}

			prefix := strings.Join(lineArgs[0:numArgs-1], " ")
			for _, cmd := range thisCmd.Subcommands {
				for _, name := range cmd.Names() {
					if !strings.HasPrefix(name, lineArgs[numArgs-1]) {
						continue
					}
					s = append(s, prefix + " " + name + " ")
				}
			}

			return
		})

		gLogs.Info(app.Description)
		gLogs.Info("提示: 方向键上下可切换历史命令.")
		gLogs.Info("提示: Ctrl + A / E 跳转命令 首 / 尾.")
		gLogs.Info("提示: 输入 help 获取帮助.")

		for {
			prompt := app.Name + " > "
			commandLine, err := line.State.Prompt(prompt)
			switch err {
			case liner.ErrPromptAborted:
				return
			case nil:
				// continue
			default:
				gLogs.Print(err)
				return
			}

			line.State.AppendHistory(commandLine)

			cmdArgs := args.Parse(commandLine)
			if len(cmdArgs) == 0 {
				continue
			}

			s := []string{os.Args[0]}
			s = append(s, cmdArgs...)

			// 恢复原始终端状态
			// 防止运行命令时程序被结束, 终端出现异常
			line.Pause()
			c.App.Run(s)
			line.Resume()
		}
	}

	app.Commands = []cli.Command{
		{
			Name:     "generate",
			Aliases:  []string{"run"},
			Usage:    "生成字典",
			Category: "生成",
			Action: func(c *cli.Context) error {

				mixPassList := make([]string, 0)

				// 开始时间
				startTime := time.Now()

				// 组合姓名
				if config.C.Storage.Name != nil && len(config.C.Storage.Name) > 0 {
					name := payload.MixName(config.C.Storage.Name)
					mixPassList = append(mixPassList, name...)
				}

				// 首字母
				if config.C.Storage.FirstLetter != "" {
					filterLetter := payload.MixFirstLetter(config.C.Storage.FirstLetter)
					mixPassList = append(mixPassList, filterLetter...)
				}

				// 组合短名称
				if config.C.Storage.Short != nil && len(config.C.Storage.Short) > 0 {
					mixPassList = append(mixPassList, config.C.Storage.Short...)
				}

				// 组合用户名
				if config.C.Storage.Username != nil && len(config.C.Storage.Username) > 0 {
					for _, v := range config.C.Storage.Username {
						username := payload.MixUsername(v)
						mixPassList = append(mixPassList, username...)
					}
				}

				// 工号
				if config.C.Storage.JobNumber != "" {
					mixPassList = append(mixPassList, config.C.Storage.JobNumber)
				}

				// QQ
				if config.C.Storage.QQ != nil && len(config.C.Storage.QQ) > 0 {
					mixPassList = append(mixPassList, config.C.Storage.QQ...)
				}

				// 组合生日
				if config.C.Storage.Birthday != "" && config.C.Storage.Lunar != "" {
					birthday := payload.MixBirthday(config.C.Storage.Birthday, config.C.Storage.Lunar)
					mixPassList = append(mixPassList, birthday...)
				}

				// 组合邮箱地址
				if config.C.Storage.Email != "" {
					email := payload.MixEmail(config.C.Storage.Email)
					mixPassList = append(mixPassList, email...)
				}

				// 组合手机号
				if config.C.Storage.Mobile != nil && len(config.C.Storage.Mobile) > 0 {
					for _, v := range config.C.Storage.Mobile {
						mobile := payload.MixMobile(v)
						mixPassList = append(mixPassList, mobile...)
					}
				}

				// 组合身份证
				if config.C.Storage.IdentityCard != "" {
					card := payload.MixIdentityCard(config.C.Storage.IdentityCard)
					mixPassList = append(mixPassList, card...)
				}

				// 组合公司/组织
				if config.C.Storage.Company != nil && len(config.C.Storage.Company) > 0 {
					company := payload.MixCompany(config.C.Storage.Company)
					mixPassList = append(mixPassList, company...)
				}

				// 组合短语
				if config.C.Storage.Phrase != "" {
					phrase := payload.MixPhrase(config.C.Storage.Phrase)
					mixPassList = append(mixPassList, phrase...)
				}

				// 组合词组
				if config.C.Storage.WordGroup != "" {
					wordGroup := payload.MixWordGroup(config.C.Storage.WordGroup)
					mixPassList = append(mixPassList, wordGroup...)
				}

				// 组合连接符
				if config.C.Storage.Connector != "" {
					connector := payload.MixConnector(config.C.Storage.Connector)
					mixPassList = append(mixPassList, connector...)
				}

				// 笛卡尔积 - 排列
				secondOrder := make([]string, 0, 10000)
				threeOrder := make([]string, 0, 15000)
				mixPassList = payload.SliceUnique(mixPassList)

				// 一阶
				firstOrder := payload.SliceUnique(append(mixPassList, payload.Pass...))
				gLogs.Infof("一阶密码生成.. %dms", time.Now().Sub(startTime) / time.Millisecond)

				// 二阶
				startTime = time.Now()
				for v := range itertools.CombinationsStr(firstOrder, 2) {
					secondOrder = append(secondOrder, strings.Join(v, ""))
				}
				gLogs.Infof("二阶密码生成.. %dms", time.Now().Sub(startTime) / time.Millisecond)

				// 三阶
				startTime = time.Now()
				for v := range itertools.CombinationsStr(firstOrder, 3) {
					threeOrder = append(threeOrder, strings.Join(v, ""))
				}
				gLogs.Infof("三阶密码生成.. %dms", time.Now().Sub(startTime) / time.Millisecond)

				total := len(firstOrder) + len(secondOrder) + len(threeOrder)
				list := make([]string, 0, total)
				list = append(list, firstOrder...)
				list = append(list, secondOrder...)
				list = append(list, threeOrder...)
				list = payload.SliceUnique(list) // 去重
				total = len(list)

				// 字典过滤
				startTime = time.Now()
				var (
					regFilterLetter *regexp.Regexp
					regFilterNumber *regexp.Regexp
				)

				// 过滤纯字符
				if config.C.Storage.FilterLetter {
					regFilterLetter, _ = regexp.Compile("^[a-zA-Z]+$")
				}
				// 过滤纯数字
				if config.C.Storage.FilterNumber {
					regFilterNumber, _ = regexp.Compile("^[0-9]+$")
				}

				dictList := make([]string, 0, total)
				for i := 0; i < total; i++ {

					// 过滤长度
					length := len([]rune(list[i]))
					if length < config.C.Storage.FilterLenMin {
						continue
					}
					if length > config.C.Storage.FilterLenMax {
						continue
					}

					// 过滤纯字符
					if regFilterLetter != nil && regFilterLetter.MatchString(list[i]) {
						continue
					}

					// 过滤纯数字
					if regFilterNumber != nil && regFilterNumber.MatchString(list[i]) {
						continue
					}

					dictList = append(dictList, list[i])
				}
				gLogs.Infof("Dict tidy.. %dms", time.Now().Sub(startTime) / time.Millisecond)

				// 输出到文件
				startTime = time.Now()

				// 文件名
				fileName := ""
				if c.IsSet("output") && c.String("output") != "" {
					fileName = c.String("output")
				} else {
					if config.C.Storage.Name == nil || len(config.C.Storage.Name) <= 0 {
						fileName = startTime.Format("2006-01-02 15:04:05") + ".txt"
					} else {
						fileName = strings.Join(config.C.Storage.Name, "") + ".txt"
					}
				}

				if err := util.OutputFile(fileName, dictList); err != nil {
					gLogs.Errorf("字典生成失败: %s", err)
				}
				gLogs.Infof("Dict output.. %dms", time.Now().Sub(startTime) / time.Millisecond)
				gLogs.Infof("Dict filename: %s", fileName)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output",
					Usage: "目标文件名",
				},
			},
		},
		{
			Name:        "filter",
			Usage:       "过滤器",
			Category:    "生成",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:        "number",
					Usage:       "过滤纯数值",
					UsageText:   app.Name + " filter number <y/n>",
					Description: `过滤纯数值`,
					Action: func(c *cli.Context) error {
						if c.NumFlags() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						if c.IsSet("y") {
							config.C.Storage.FilterNumber = true
						}

						if c.IsSet("n") {
							config.C.Storage.FilterNumber = false
						}

						gLogs.Debugf("过滤纯数值: %s", strconv.FormatBool(config.C.Storage.FilterNumber))
						return nil
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "y",
							Usage: "是",
						},
						cli.BoolFlag{
							Name:  "n",
							Usage: "否",
						},
					},
				},
				{
					Name:        "letter",
					Usage:       "过滤纯字母",
					UsageText:   app.Name + " filter letter <y/n>",
					Description: `过滤纯字母`,
					Action: func(c *cli.Context) error {
						if c.NumFlags() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						if c.IsSet("y") {
							config.C.Storage.FilterLetter = true
						}

						if c.IsSet("n") {
							config.C.Storage.FilterLetter = false
						}

						gLogs.Debugf("过滤纯字母: %s", strconv.FormatBool(config.C.Storage.FilterLetter))
						return nil
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "y",
							Usage: "是",
						},
						cli.BoolFlag{
							Name:  "n",
							Usage: "否",
						},
					},
				},
				{
					Name:        "min",
					Usage:       "过滤长度 - 最小值",
					UsageText:   app.Name + " filter min <最小值>",
					Description: `过滤长度小于min的密码`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						min, err := strconv.Atoi(c.Args().Get(0))
						if err != nil {
							gLogs.Warn("数值错误")
							return nil
						}

						config.C.Storage.FilterLenMin = min

						gLogs.Debugf("过滤长度最小值: %s", strconv.Itoa(config.C.Storage.FilterLenMin))
						return nil
					},
				},
				{
					Name:        "max",
					Usage:       "过滤长度 - 最大值",
					UsageText:   app.Name + " filter max <最大值>",
					Description: `过滤长度大于max的密码`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						max, err := strconv.Atoi(c.Args().Get(0))
						if err != nil {
							gLogs.Warn("数值错误")
							return nil
						}

						config.C.Storage.FilterLenMax = max

						gLogs.Debugf("过滤长度最大值: %s", strconv.Itoa(config.C.Storage.FilterLenMax))
						return nil
					},
				},
			},
		},
		{
			Name:     "set",
			Usage:    "设置属性",
			Category: "生成",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:        "name",
					Usage:       "姓名(中文/英文)",
					UsageText:   app.Name + " set name <姓名(周杰伦/zhou jie lun)>",
					Description: `姓名支持(中文/英文)输入, 每个字之间需空格分隔`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// 纯中文正则
						reg, err := regexp.Compile("^[\u4e00-\u9fa5]+$")
						if err != nil {
							panic(err)
						}

						// 纯拼音正则
						enReg, err := regexp.Compile("^[a-zA-Z]+$")
						if err != nil {
							panic(err)
						}

						params := c.Args()

						// 设置姓名
						if reg.MatchString(params[0]) {
							// 若为纯中文，则转换为拼音
							config.C.Storage.Name = pinyin.ConvertNameSlice(params[0])
						} else if enReg.MatchString(params[0]) {
							// 全拼音
							config.C.Storage.Name = pinyin.FormatSliceToLower(params)
						} else {
							gLogs.Warn("姓名格式错误，请输入纯中文或拼音")
							return nil
						}

						// 设置首字母
						if config.C.Storage.Name != nil {
							config.C.Storage.FirstLetter = pinyin.FormatSliceFirstLetter(config.C.Storage.Name)
						}

						gLogs.Debugf("姓名: %s", strings.Join(config.C.Storage.Name, " "))
						gLogs.Debugf("首字母: %s", config.C.Storage.FirstLetter)
						return nil
					},
				},
				{
					Name:        "short",
					Usage:       "短名称(英文)",
					UsageText:   app.Name + " set short <短名称(zhoujl)>",
					Description: `短名称(英文)`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						config.C.Storage.Short = []string(c.Args())

						gLogs.Debugf("短名称: %s", strings.Join(config.C.Storage.Short, " "))
						return nil
					},
				},
				{
					Name:        "first",
					Usage:       "姓名首字母(英文)",
					UsageText:   app.Name + " set first <姓名首字母(zjl)>",
					Description: `姓名首字母(英文),默认自动获取姓名首字母`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						params := c.Args()
						config.C.Storage.FirstLetter = strings.ToLower(params[0])

						gLogs.Debugf("首字母: %s", config.C.Storage.FirstLetter)
						return nil
					},
				},
				{
					Name:        "birthday",
					Usage:       "公历生日(yyyymmdd)",
					UsageText:   app.Name + " set birthday <公历生日(yyyymmdd)>",
					Description: `公历生日,格式：yyyymmdd. 默认根据身份证自动计算`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						ymd := c.Args().Get(0)

						t, err := time.Parse("20060504", ymd)
						if err != nil {
							gLogs.Error("公历生日转换失败")
							return nil
						}

						// 设置公历生日
						config.C.Storage.Birthday = t.Format("20060504")

						// 公历转换农历生日
						str, _ := solarlunar.SolarToLuanr(t.Format("2006-05-04"))
						t, err = time.Parse("2006-05-04", str)
						if err == nil {
							config.C.Storage.Lunar = t.Format("20060504")
						}

						gLogs.Debugf("公历生日: %s", config.C.Storage.Birthday)
						gLogs.Debugf("农历生日: %s", config.C.Storage.Lunar)
						return nil
					},
				},
				{
					Name:        "lunar",
					Usage:       "农历生日(yyyymmdd)",
					UsageText:   app.Name + " set lunar <农历生日(yyyymmdd)>",
					Description: `农历生日,格式：yyyymmdd. 默认根据公历自动计算`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						ymd := c.Args().Get(0)

						t, err := time.Parse("20060504", ymd)
						if err != nil {
							gLogs.Error("农历生日转换失败")
							return nil
						}

						// 设置农历生日
						config.C.Storage.Lunar = t.Format("20060504")

						// 农历转换公历生日
						str := solarlunar.LunarToSolar(t.Format("2006-05-04"), false)
						t, err = time.Parse("2006-05-04", str)
						if err == nil {
							config.C.Storage.Birthday = t.Format("20060504")
						}

						gLogs.Debugf("农历生日: %s", config.C.Storage.Lunar)
						gLogs.Debugf("公历生日: %s", config.C.Storage.Birthday)
						return nil
					},
				},
				{
					Name:        "email",
					Usage:       "邮箱地址",
					UsageText:   app.Name + " set email <邮箱地址>",
					Description: `邮箱地址`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						reg, err := regexp.Compile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`);
						if err != nil {
							panic(err)
						}

						email := c.Args().Get(0)
						if !reg.MatchString(email) {
							gLogs.Error("您输入的邮箱格式不正确")
							return nil
						}

						// 设置邮箱地址
						config.C.Storage.Email = email

						gLogs.Debugf("邮箱: %s", config.C.Storage.Email)
						return nil
					},
				},
				{
					Name:        "mobile",
					Usage:       "手机号码",
					UsageText:   app.Name + " set mobile <手机号码>",
					Description: `手机号码`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						reg, err := regexp.Compile(`^1[0-9]{10}$`);
						if err != nil {
							panic(err)
						}

						mobileList := make([]string, 0)
						for _, mobile := range c.Args() {
							if !reg.MatchString(mobile) {
								gLogs.Warnf("您输入的手机号码格式不正确: %s", mobile)
								continue
							}

							mobileList = append(mobileList, mobile)
						}

						// 设置手机号码
						config.C.Storage.Mobile = mobileList

						gLogs.Debugf("手机号码: %s", strings.Join(config.C.Storage.Mobile, " "))
						return nil
					},
				},
				{
					Name:        "username",
					Usage:       "用户名(英文)",
					UsageText:   app.Name + " set username <用户名(英文)>",
					Description: `用户名(英文)`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// 设置用户名
						config.C.Storage.Username = []string(c.Args())

						gLogs.Debugf("用户名: %s", strings.Join(config.C.Storage.Username, " "))
						return nil
					},
				},
				{
					Name:        "qq",
					Usage:       "腾讯QQ",
					UsageText:   app.Name + " set qq <腾讯QQ>",
					Description: `腾讯QQ号码`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// 设置QQ
						config.C.Storage.QQ = []string(c.Args())

						gLogs.Debugf("QQ: %s", strings.Join(config.C.Storage.QQ, " "))
						return nil
					},
				},
				{
					Name:        "company",
					Usage:       "企业/组织",
					UsageText:   app.Name + " set company <企业/组织>",
					Description: `企业/组织`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// 纯中文正则
						reg, err := regexp.Compile("^[\u4e00-\u9fa5]+$")
						if err != nil {
							panic(err)
						}

						// 是否中文
						if c.NArg() == 1 && reg.MatchString(c.Args().Get(0)) {
							company := c.Args().Get(0)
							config.C.Storage.Company = pinyin.ConvertSlice(company)
						} else {
							config.C.Storage.Company = []string(c.Args())
						}

						// 首字母
						companyFirstLetter := pinyin.FormatSliceFirstLetter(config.C.Storage.Company)

						gLogs.Debugf("企业/组织: %s", strings.Join(config.C.Storage.Company, " "))
						gLogs.Debugf("首字母: %s", companyFirstLetter)
						return nil
					},
				},
				{
					Name:        "phrase",
					Usage:       "英文短语",
					UsageText:   app.Name + " set phrase <短语/woaini>",
					Description: `英文短语(iloveyou、woaini)`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// 设置英文短语
						phrase := c.Args().Get(0)
						config.C.Storage.Phrase = phrase

						gLogs.Debugf("英文短语: %s", config.C.Storage.Phrase)
						return nil
					},
				},
				{
					Name:        "card",
					Usage:       "身份证",
					UsageText:   app.Name + " set card <身份证>",
					Description: `身份证号码(18位)`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						reg, err := regexp.Compile(`^[1-9]\d{7}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}$|^[1-9]\d{5}[1-9]\d{3}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}([0-9]|[xX])$`);
						if err != nil {
							panic(err)
						}

						card := c.Args().Get(0)
						if !reg.MatchString(card) {
							gLogs.Error("您输入的身份证号码格式不正确")
							return nil
						}

						// 设置身份证
						config.C.Storage.IdentityCard = card

						gLogs.Debugf("身份证: %s", config.C.Storage.IdentityCard)
						return nil
					},
				},
				{
					Name:        "no",
					Usage:       "工号",
					UsageText:   app.Name + " set no <工号>",
					Description: `工号`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						job := c.Args().Get(0)
						config.C.Storage.JobNumber = job

						gLogs.Debugf("工号: %s", config.C.Storage.JobNumber)
						return nil
					},
				},
				{
					Name:        "word",
					Usage:       "常用词组",
					UsageText:   app.Name + " set word <常用词组>",
					Description: `常用词组,逗号分隔`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						wordGroup := c.Args().Get(0)
						config.C.Storage.WordGroup = wordGroup

						gLogs.Debugf("常用词组: %s", config.C.Storage.WordGroup)
						return nil
					},
				},
				{
					Name:        "connector",
					Usage:       "连接符",
					UsageText:   app.Name + " set connector <连接符>",
					Description: `连接符. 默认 @#.-_~!?%&*+=$/|`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						connector := c.Args().Get(0)
						config.C.Storage.Connector = connector

						gLogs.Debugf("连接符: %s", config.C.Storage.Connector)
						return nil
					},
				},
			},
		},
		{
			Name:        "print",
			Aliases:     []string{"p"},
			Usage:       "打印",
			UsageText:   app.Name + " print",
			Description: "打印信息",
			Category:    "其他",
			Action: func(c *cli.Context) error {

				data := [][]string{
					[]string{"姓名", "name", strings.Join(config.C.Storage.Name, " ")},
					[]string{"首字母", "first", config.C.Storage.FirstLetter},
					[]string{"短名称", "short", strings.Join(config.C.Storage.Short, " ")},
					[]string{"用户名", "username", strings.Join(config.C.Storage.Username, " ")},
					[]string{"手机号", "mobile", strings.Join(config.C.Storage.Mobile, " ")},
					[]string{"QQ", "qq", strings.Join(config.C.Storage.QQ, " ")},
					[]string{"邮箱", "email", config.C.Storage.Email},
					[]string{"工号", "no", config.C.Storage.JobNumber},
					[]string{"公历生日", "birthday", config.C.Storage.Birthday},
					[]string{"农历生日", "lunar", config.C.Storage.Lunar},
					[]string{"身份证", "card", config.C.Storage.IdentityCard},
					[]string{"公司/组织", "company", strings.Join(config.C.Storage.Company, " ")},
					[]string{"短语", "phrase", config.C.Storage.Phrase},
					[]string{"常用词组", "word", config.C.Storage.WordGroup},
					[]string{"连接符", "connector", config.C.Storage.Connector},
					[]string{"是否过滤纯数字", "filter number", strconv.FormatBool(config.C.Storage.FilterNumber)},
					[]string{"是否过滤纯字母", "filter letter", strconv.FormatBool(config.C.Storage.FilterLetter)},
					[]string{"过滤长度 - min", "filter min", strconv.Itoa(config.C.Storage.FilterLenMin)},
					[]string{"过滤长度 - max", "filter max", strconv.Itoa(config.C.Storage.FilterLenMax)},
				}

				t := table.NewTable(os.Stdout)
				t.AppendBulk(data)
				t.Render()
				return nil
			},
		},
		{
			Name:        "reset",
			Usage:       "重置",
			UsageText:   app.Name + " reset <待重置参数>",
			Description: "重置默认设置",
			Category:    "其他",
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					// 重置所有属性的默认设置
					if err := config.C.Reset(); err != nil {
						gLogs.Warnf("重置失败: %s", err)
					}
					return nil
				}

				// 重置单独属性
				switch c.Args().Get(0) {
				case "name":
					config.C.Storage.Name = make([]string, 0)
				case "first":
					config.C.Storage.FirstLetter = ""
				case "short":
					config.C.Storage.Short = make([]string, 0)
				case "username":
					config.C.Storage.Username = make([]string, 0)
				case "birthday":
					config.C.Storage.Birthday = ""
				case "lunar":
					config.C.Storage.Lunar = ""
				case "email":
					config.C.Storage.Email = ""
				case "mobile":
					config.C.Storage.Mobile = make([]string, 0)
				case "qq":
					config.C.Storage.QQ = make([]string, 0)
				case "company":
					config.C.Storage.Company = make([]string, 0)
				case "phrase":
					config.C.Storage.Phrase = config.Phrase
				case "card":
					config.C.Storage.IdentityCard = ""
				case "no":
					config.C.Storage.JobNumber = ""
				case "word":
					config.C.Storage.WordGroup = config.WordGroup
				case "connector":
					config.C.Storage.Connector = config.Connector
				default:
					gLogs.Warn("未找到该属性")
				}
				return nil
			},
		},
		{
			Name:        "clear",
			Aliases:     []string{"cls"},
			Usage:       "清空控制台",
			UsageText:   app.Name + " clear",
			Description: "清空控制台屏幕",
			Category:    "其他",
			Action: func(c *cli.Context) error {
				dictliner.ClearScreen()
				return nil
			},
		},
		{
			Name:    "quit",
			Aliases: []string{"q", "exit"},
			Usage:   "退出",
			Action: func(c *cli.Context) error {
				return cli.NewExitError("", 0)
			},
			Hidden:   true,
			HideHelp: true,
		},
	}

	if err := app.Run(os.Args); err != nil {
		gLogs.Panic(err.Error())
	}
}
