<h1 align="center">
  <br>DictGenerate<br>
</h1>

# 📖 Introduction

使用Go语言编写的社工字典生成器

The social engineering dictionary generator written by Go

[![Build Status](https://travis-ci.org/telanflow/DictGenerate.svg?branch=master)](https://travis-ci.org/telanflow/DictGenerate)
[![Go version](https://img.shields.io/badge/Go-1.13-brightgreen.svg)](https://github.com/telanflow/DictGenerate)
[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)
[![LICENSE](https://img.shields.io/badge/license-NPL%20(The%20996%20Prohibited%20License)-blue.svg)](https://github.com/996icu/996.ICU/blob/master/LICENSE)

## 🚀 Feature
- 支持 姓名/用户名/首字母/公司(组织)/邮箱/QQ/手机号/身份证/生日(公历农历)/工号/常用词/连接符/短语/弱口令 等
- 支持多种密码组合混淆方式
- 支持`中文自动转拼音` / `姓名自动提取首字母` / `生日自动转换`（公历/农历）/ `身份证自动转换`
- 配置信息留存 / 历史命令 / tab智能提示
- 多阶段密码字典生成 简易(easy) / 中等(medium) / 复杂(large)

## 🧰 Install
```
git clone git@github.com:telanflow/DictGenerate.git
cd DictGenerate
sh ./build.sh
```
[Download](https://github.com/telanflow/DictGenerate/releases)

## 🛠 Usage
    USAGE:
       DictGenerate [global options] command [command options] [arguments...]
    
    COMMANDS:
         help, h  Shows a list of commands or help for one command
       其他:
         print, p    打印
         reset       重置
         clear, cls  清空控制台
       生成:
         generate, run  生成字典
         filter         过滤器
         set            设置属性
             name       姓名(中文/英文)
             short      短名称(英文)
             first      姓名首字母(英文)
             birthday   公历生日(yyyymmdd)
             lunar      农历生日(yyyymmdd)
             email      邮箱地址
             mobile     手机号码
             username   用户名(英文)
             qq         腾讯QQ
             company    企业/组织
             phrase     英文短语
             card       身份证
             no         工号
             word       常用词组
             connector  连接符
    
    GLOBAL OPTIONS:
       --help, -h     show help
       --version, -v  print the version
    
    COPYRIGHT:
       (c) 2019 telanflow.

## Screenshot

![screenshot](screenshot.png)

## 📄 License
Source code in `DictGenerate` is available under the [BSD 3 License](/LICENSE).