package payload

import (
	"testing"
)

func TestSliceUnique(t *testing.T) {
	s := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"E",
		"f",
		"F",

		"f",
		"F",
	}

	u := SliceUnique(s)
	if len(u) != 8 {
		t.Fatal("fail")
	}
}

func TestReverseString(t *testing.T) {
	s := "abcd"
	rs := ReverseString(s)
	if rs != "dcba" {
		t.Fail()
	}
}

func TestMixName(t *testing.T) {
	s := []string{
		"zhou",
		"jie",
		"lun",
	}

	resp := MixName(s)
	for _, v := range resp {
		t.Log(v)
	}
}

func TestMixBirthday(t *testing.T) {
	birthday := "19950715"
	lunar := "19950618"

	resp := MixBirthday(birthday, lunar)
	for _, v := range resp {
		t.Log(v)
	}
}

func TestMixEmail(t *testing.T) {
	email := "zhoujielun@qq.com"
	resp := MixEmail(email)
	for _, v := range resp {
		t.Log(v)
	}
}

func TestMixFirstLetter(t *testing.T) {
	s := "zjl"

	resp := MixFirstLetter(s)
	for _, v := range resp {
		t.Log(v)
	}
}

func TestMixMobile(t *testing.T) {
	mobile := "13077870989"
	resp := MixMobile(mobile)
	for _, v := range resp {
		t.Log(v)
	}
}