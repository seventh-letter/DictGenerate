package main

import (
	"DictGenerate/inter/config"
	"DictGenerate/inter/dictliner"
	"DictGenerate/inter/dictliner/args"
	"DictGenerate/inter/logger"
	"DictGenerate/inter/payload"
	"DictGenerate/inter/pinyin"
	"DictGenerate/inter/table"
	"DictGenerate/util"
	"fmt"
	"github.com/ernestosuarez/itertools"
	"github.com/nosixtools/solarlunar"
	"github.com/peterh/liner"
	"github.com/urfave/cli"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	reloadFn = func(c *cli.Context) error {
		err := config.C.Reload()
		if err != nil {
			logger.Warnf("重载配置错误: %s", err)
		}
		return nil
	}
	saveFunc = func(c *cli.Context) error {
		err := config.C.Save()
		if err != nil {
			logger.Warnf("保存配置错误: %s", err)
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
		logger.Fatalf("配置文件无权限访问：%s", config.ConfigFilePath)
	case config.ErrConfigContentsParseError:
		logger.Error("Resolve config fail, reset it!")
		logger.Info("Reset config ...")
		logger.Infof("config path: %s", config.ConfigFilePath)
		err := config.C.Reset()
		if err != nil {
			logger.Fatalf("Reset config fail! 请尝试手动删除配置文件：%s", config.ConfigFilePath)
		}

		logger.Info("Reload config ...")
		if err := config.C.Init(); err != nil {
			logger.Fatal("Reload config fail! ")
		}

		logger.Info("Config loaded successfully!")

	default:
		fmt.Printf("WARNING: config init error: %s\n", err)
	}
}

func main() {
	defer config.C.Close()

	app := cli.NewApp()
	app.Name = config.AppName
	app.Author = config.Author
	app.Version = config.Version
	app.Email = config.Email
	app.Usage = app.Name + " for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Copyright = "(c) 2019 " + app.Author + "."
	app.Description = config.NameChar + `
使用Go语言编写的社工字典生成器`

	app.Action = func(c *cli.Context) {
		if c.NArg() != 0 {
			logger.Warnf("未找到命令: %s", c.Args().Get(0))
			logger.Warnf("运行命令 %s help 获取帮助", app.Name)
			return
		}

		var (
			line = dictliner.NewLiner()
			err  error
		)

		line.History, err = dictliner.NewLineHistory(config.HistoryFilePath)
		if err != nil {
			logger.Warnf("警告: 读取历史命令文件错误, %s", err)
		}

		_ = line.ReadHistory()
		defer func() {
			_ = line.DoWriteHistory()
			_ = line.Close()
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
					s = append(s, prefix+" "+name+" ")
				}
			}

			return
		})

		fmt.Println(app.Description)
		fmt.Println("提示: 方向键上下可切换历史命令.")
		fmt.Println("提示: Ctrl + A / E 跳转命令 首 / 尾.")
		fmt.Println("提示: 输入 help 获取帮助.")
		fmt.Println("")

		for {
			prompt := app.Name + " > "
			commandLine, err := line.State.Prompt(prompt)
			switch err {
			case liner.ErrPromptAborted:
				return
			case nil:
				// continue
			default:
				logger.Error(err)
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
				nameList := make([]string, 0)
				if config.C.Storage.Name != nil && len(config.C.Storage.Name) > 0 {
					nameList = payload.MixName(config.C.Storage.Name)
					mixPassList = append(mixPassList, nameList...)
				}

				// 首字母
				filterLetterList := make([]string, 0)
				if config.C.Storage.FirstLetter != "" {
					filterLetterList = payload.MixFirstLetter(config.C.Storage.FirstLetter)
					mixPassList = append(mixPassList, filterLetterList...)
				}

				// 组合短名称
				shortNameList := make([]string, 0)
				if config.C.Storage.Short != nil && len(config.C.Storage.Short) > 0 {
					shortNameList = config.C.Storage.Short
					mixPassList = append(mixPassList, shortNameList...)
				}

				// 组合用户名
				usernameList := make([]string, 0)
				if config.C.Storage.Username != nil && len(config.C.Storage.Username) > 0 {
					for _, v := range config.C.Storage.Username {
						username := payload.MixUsername(v)
						usernameList = append(usernameList, username...)
					}
					mixPassList = append(mixPassList, usernameList...)
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
				birthdayList := make([]string, 0)
				if config.C.Storage.Birthday != "" && config.C.Storage.Lunar != "" {
					birthdayList = payload.MixBirthday(config.C.Storage.Birthday, config.C.Storage.Lunar)
					mixPassList = append(mixPassList, birthdayList...)
				}

				// 组合邮箱地址
				emailList := make([]string, 0)
				if config.C.Storage.Email != nil && len(config.C.Storage.Email) > 0 {
					for _, v := range config.C.Storage.Email {
						email := payload.MixEmail(v)
						emailList = append(emailList, email...)
					}
					mixPassList = append(mixPassList, emailList...)
				}

				// 组合手机号
				mobileList := make([]string, 0)
				if config.C.Storage.Mobile != nil && len(config.C.Storage.Mobile) > 0 {
					for _, v := range config.C.Storage.Mobile {
						mobile := payload.MixMobile(v)
						mobileList = append(mobileList, mobile...)
					}
					mixPassList = append(mixPassList, mobileList...)
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
				connectorList := make([]string, 0)
				if config.C.Storage.Connector != "" {
					connectorList = payload.MixConnector(config.C.Storage.Connector)
					mixPassList = append(mixPassList, connectorList...)
				}

				// 组合列表
				combinationList := make([]string, 0)
				// 姓名&连接符&生日
				if len(nameList) > 0 && len(birthdayList) > 0 {
					list := make([]string, 0, len(nameList)+len(birthdayList))
					list = append(list, nameList...)
					list = append(list, birthdayList...)

					mixList := make([]string, 0, len(list))
					for v := range itertools.CombinationsStr(list, 2) {
						for _, connector := range connectorList {
							mixList = append(mixList, strings.Join(v, connector))
						}
					}
					combinationList = append(combinationList, mixList...)
				}
				// 用户名&连接符&生日
				if len(usernameList) > 0 && len(birthdayList) > 0 {
					list := make([]string, 0, len(usernameList)+len(birthdayList))
					list = append(list, usernameList...)
					list = append(list, birthdayList...)

					mixList := make([]string, 0, len(list))
					for v := range itertools.CombinationsStr(list, 2) {
						for _, connector := range connectorList {
							mixList = append(mixList, strings.Join(v, connector))
						}
					}
					combinationList = append(combinationList, mixList...)
				}
				// 短名称&连接符&生日
				if len(shortNameList) > 0 && len(birthdayList) > 0 {
					list := make([]string, 0, len(shortNameList)+len(birthdayList))
					list = append(list, shortNameList...)
					list = append(list, birthdayList...)

					mixList := make([]string, 0, len(list))
					for v := range itertools.CombinationsStr(list, 2) {
						for _, connector := range connectorList {
							mixList = append(mixList, strings.Join(v, connector))
						}
					}
					combinationList = append(combinationList, mixList...)
				}
				// 姓名&连接符&手机号
				if len(nameList) > 0 && len(mobileList) > 0 {
					list := make([]string, 0, len(nameList)+len(mobileList))
					list = append(list, nameList...)
					list = append(list, mobileList...)

					mixList := make([]string, 0, len(list))
					for v := range itertools.CombinationsStr(list, 2) {
						for _, connector := range connectorList {
							mixList = append(mixList, strings.Join(v, connector))
						}
					}
					combinationList = append(combinationList, mixList...)
				}
				// 短名称&连接符&手机号
				if len(shortNameList) > 0 && len(mobileList) > 0 {
					list := make([]string, 0, len(shortNameList)+len(mobileList))
					list = append(list, shortNameList...)
					list = append(list, mobileList...)

					mixList := make([]string, 0, len(list))
					for v := range itertools.CombinationsStr(list, 2) {
						for _, connector := range connectorList {
							mixList = append(mixList, strings.Join(v, connector))
						}
					}
					combinationList = append(combinationList, mixList...)
				}

				// 笛卡尔积 - 排列
				mixPassList = payload.SliceUnique(append(mixPassList, payload.Pass...))

				// 文件名
				fileName := ""
				if c.IsSet("output") && c.String("output") != "" {
					fileName = c.String("output")
				} else {
					if config.C.Storage.Name == nil || len(config.C.Storage.Name) <= 0 {
						fileName = startTime.Format("2006-01-02 15:04:05")
					} else {
						fileName = strings.Join(config.C.Storage.Name, "")
					}
				}

				// 并发生成
				logger.Info("Dict generate..")
				wg := sync.WaitGroup{}
				wg.Add(3)

				// 一阶
				go func() {
					file := fmt.Sprintf("%s_%s", fileName, "easy")

					firstOrder := make([]string, 0, len(mixPassList)+len(combinationList))
					firstOrder = append(firstOrder, mixPassList...)
					firstOrder = append(firstOrder, combinationList...)
					firstOrder = payload.SliceUnique(firstOrder)
					if filePath, err := DictOutput(firstOrder, file); err != nil {
						logger.Errorf("easy generate fail: %s", err)
					} else {
						logger.Infof("easy dict filename: %s", filePath)
					}
					wg.Done()
				}()

				// 二阶
				go func() {
					file := fmt.Sprintf("%s_%s", fileName, "medium")

					secondOrder := make([]string, 0, len(mixPassList)*2)
					for v := range itertools.CombinationsStr(mixPassList, 2) {
						secondOrder = append(secondOrder, strings.Join(v, ""))
					}

					secondOrder = append(secondOrder, combinationList...)
					if filePath, err := DictOutput(secondOrder, file); err != nil {
						logger.Errorf("medium generate fail: %s", err)
					} else {
						logger.Infof("medium dict filename: %s", filePath)
					}
					wg.Done()
				}()

				// 三阶
				go func() {
					file := fmt.Sprintf("%s_%s", fileName, "large")

					threeOrder := make([]string, 0, len(mixPassList)*2)
					for v := range itertools.CombinationsStr(mixPassList, 3) {
						threeOrder = append(threeOrder, strings.Join(v, ""))
					}

					threeOrder = append(threeOrder, combinationList...)
					if filePath, err := DictOutput(threeOrder, file); err != nil {
						logger.Errorf("large generate fail: %s", err)
					} else {
						logger.Infof("large dict filename: %s", filePath)
					}
					wg.Done()
				}()

				wg.Wait()

				logger.Infof("OK Generate completed!")

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
			Name:     "filter",
			Usage:    "过滤器",
			Category: "生成",
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

						logger.Infof("过滤纯数值: %s", strconv.FormatBool(config.C.Storage.FilterNumber))
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

						logger.Infof("过滤纯字母: %s", strconv.FormatBool(config.C.Storage.FilterLetter))
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
							logger.Warn("数值错误")
							return nil
						}

						config.C.Storage.FilterLenMin = min

						logger.Infof("过滤长度最小值: %s", strconv.Itoa(config.C.Storage.FilterLenMin))
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
							logger.Warn("数值错误")
							return nil
						}

						config.C.Storage.FilterLenMax = max

						logger.Infof("过滤长度最大值: %s", strconv.Itoa(config.C.Storage.FilterLenMax))
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
						reg := regexp.MustCompile("^[\u4e00-\u9fa5]+$")
						// 纯拼音正则
						enReg := regexp.MustCompile("^[a-zA-Z]+$")

						params := c.Args()

						// 设置姓名
						if reg.MatchString(params[0]) {
							// 若为纯中文，则转换为拼音
							config.C.Storage.Name = pinyin.ConvertNameSlice(params[0])
						} else if enReg.MatchString(params[0]) {
							// 全拼音
							config.C.Storage.Name = pinyin.FormatSliceToLower(params)
						} else {
							logger.Warn("姓名格式错误，请输入纯中文或拼音")
							return nil
						}

						// 设置首字母
						if config.C.Storage.Name != nil {
							config.C.Storage.FirstLetter = pinyin.FormatSliceFirstLetter(config.C.Storage.Name)
						}

						logger.Infof("姓名: %s", strings.Join(config.C.Storage.Name, " "))
						logger.Infof("首字母: %s", config.C.Storage.FirstLetter)
						return nil
					},
				},
				{
					Name:        "short",
					Usage:       "短名称(英文)",
					UsageText:   app.Name + " set short <短名称(zhoujl zhoujl)> ",
					Description: `短名称(英文) 支持添加多个，空格区分`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						config.C.Storage.Short = c.Args()
						logger.Infof("短名称: %s", strings.Join(config.C.Storage.Short, " "))
						return nil
					},
				},
				{
					Name:        "first",
					Usage:       "姓名首字母(英文: zjl)",
					UsageText:   app.Name + " set first <姓名首字母(zjl)>",
					Description: `姓名首字母(英文),默认自动获取姓名首字母`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						params := c.Args()
						config.C.Storage.FirstLetter = strings.ToLower(params[0])

						logger.Infof("姓名首字母: %s", config.C.Storage.FirstLetter)
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
							logger.Errorf("公历生日格式错误: %s", err)
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

						logger.Infof("公历生日: %s", config.C.Storage.Birthday)
						logger.Infof("农历生日: %s", config.C.Storage.Lunar)
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
							logger.Errorf("农历生日格式错误: %s", err)
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

						logger.Infof("农历生日: %s", config.C.Storage.Lunar)
						logger.Infof("公历生日: %s", config.C.Storage.Birthday)
						return nil
					},
				},
				{
					Name:        "email",
					Usage:       "邮箱地址",
					UsageText:   app.Name + " set email <邮箱地址(xxx@gmail.com xxx@qq.com)>",
					Description: `邮箱地址 支持多个，空格区分`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						reg := regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)

						emailList := make([]string, 0)
						for _, email := range c.Args() {
							if email == "" {
								continue
							}

							if !reg.MatchString(email) {
								logger.Errorf("您输入的邮箱格式不正确: %s", email)
								continue
							}

							emailList = append(emailList, email)
						}

						// 设置邮箱地址
						config.C.Storage.Email = emailList

						logger.Infof("邮箱: %s", strings.Join(config.C.Storage.Email, " "))
						return nil
					},
				},
				{
					Name:        "mobile",
					Usage:       "手机号码",
					UsageText:   app.Name + " set mobile <手机号码(13011111111 15622222222)>",
					Description: `手机号码 支持多个 空格区分`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						reg := regexp.MustCompile(`^1[0-9]{10}$`)

						mobileList := make([]string, 0)
						for _, mobile := range c.Args() {
							if !reg.MatchString(mobile) {
								logger.Errorf("您输入的手机号码格式不正确: %s", mobile)
								continue
							}

							mobileList = append(mobileList, mobile)
						}

						// 设置手机号码
						config.C.Storage.Mobile = mobileList

						logger.Infof("手机号码: %s", strings.Join(config.C.Storage.Mobile, " "))
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
						config.C.Storage.Username = c.Args()

						logger.Infof("用户名: %s", strings.Join(config.C.Storage.Username, " "))
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
						config.C.Storage.QQ = c.Args()

						logger.Infof("QQ: %s", strings.Join(config.C.Storage.QQ, " "))
						return nil
					},
				},
				{
					Name:        "company",
					Usage:       "企业/组织",
					UsageText:   app.Name + " set company <企业/组织> 中文自动转拼音",
					Description: `企业/组织`,
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// 纯中文正则
						reg := regexp.MustCompile("^[\u4e00-\u9fa5]+$")

						// 是否中文
						if c.NArg() == 1 && reg.MatchString(c.Args().Get(0)) {
							company := c.Args().Get(0)
							config.C.Storage.Company = pinyin.ConvertSlice(company)
						} else {
							config.C.Storage.Company = []string(c.Args())
						}

						// 首字母
						companyFirstLetter := pinyin.FormatSliceFirstLetter(config.C.Storage.Company)

						logger.Infof("企业/组织: %s", strings.Join(config.C.Storage.Company, " "))
						logger.Infof("首字母: %s", companyFirstLetter)
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

						logger.Infof("英文短语: %s", config.C.Storage.Phrase)
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

						reg := regexp.MustCompile(`^[1-9]\d{7}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}$|^[1-9]\d{5}[1-9]\d{3}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}([0-9]|[xX])$`)

						card := c.Args().Get(0)
						if !reg.MatchString(card) {
							logger.Error("您输入的身份证号码格式不正确")
							return nil
						}

						// 设置身份证
						config.C.Storage.IdentityCard = card

						logger.Infof("身份证: %s", config.C.Storage.IdentityCard)
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

						logger.Infof("工号: %s", config.C.Storage.JobNumber)
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

						logger.Infof("常用词组: %s", config.C.Storage.WordGroup)
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

						logger.Infof("连接符: %s", config.C.Storage.Connector)
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
					{"姓名", "name", strings.Join(config.C.Storage.Name, " ")},
					{"首字母", "first", config.C.Storage.FirstLetter},
					{"短名称", "short", strings.Join(config.C.Storage.Short, " ")},
					{"用户名", "username", strings.Join(config.C.Storage.Username, " ")},
					{"手机号", "mobile", strings.Join(config.C.Storage.Mobile, " ")},
					{"QQ", "qq", strings.Join(config.C.Storage.QQ, " ")},
					{"邮箱", "email", strings.Join(config.C.Storage.Email, " ")},
					{"工号", "no", config.C.Storage.JobNumber},
					{"公历生日", "birthday", config.C.Storage.Birthday},
					{"农历生日", "lunar", config.C.Storage.Lunar},
					{"身份证", "card", config.C.Storage.IdentityCard},
					{"公司/组织", "company", strings.Join(config.C.Storage.Company, " ")},
					{"短语", "phrase", config.C.Storage.Phrase},
					{"常用词组", "word", config.C.Storage.WordGroup},
					{"连接符", "connector", config.C.Storage.Connector},
					{"是否过滤纯数字", "filter number", strconv.FormatBool(config.C.Storage.FilterNumber)},
					{"是否过滤纯字母", "filter letter", strconv.FormatBool(config.C.Storage.FilterLetter)},
					{"过滤长度 - min", "filter min", strconv.Itoa(config.C.Storage.FilterLenMin)},
					{"过滤长度 - max", "filter max", strconv.Itoa(config.C.Storage.FilterLenMax)},
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
						logger.Warnf("重置失败: %s", err)
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
					config.C.Storage.Email = make([]string, 0)
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
					logger.Warn("未找到该属性")
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
		logger.Fatal(err.Error())
	}
}

// DictOutput 输出字典列表到文件
func DictOutput(list []string, fileName string) (filePath string, err error) {
	var (
		regFilterLetter *regexp.Regexp
		regFilterNumber *regexp.Regexp
		total           = len(list)
	)

	// 过滤纯字符
	if config.C.Storage.FilterLetter {
		regFilterLetter = regexp.MustCompile("^[a-zA-Z]+$")
	}
	// 过滤纯数字
	if config.C.Storage.FilterNumber {
		regFilterNumber = regexp.MustCompile("^[0-9]+$")
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

	// 生成文件
	filePath = fmt.Sprintf("%s.txt", fileName)
	err = util.OutputFile(filePath, dictList)
	return
}
