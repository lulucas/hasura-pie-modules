package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/schema"
	"net/url"
	"sort"
	"strings"
)

type SignFunc func(interface{}) (string, error)

type SignOption struct {
	// 为不参与签名的字段名称 默认空
	IgnoreKeys []string
	// 是否跳过空值 默认否
	IgnoreEmpty bool
	// 传入键值，返回拼接结果 默认{key}={value}
	KeyValueFunc func(key, value string) string
	// 签名字符串后处理，通常用于添加密钥
	PostSignStrHook func(string) string
	// 拼接符号 默认&
	JoinSep string
	// 哈希算法 默认MD5
	HashFunc func(string) string
}

// 通用签名函数
func NewSignFunc(opt SignOption) SignFunc {
	if opt.KeyValueFunc == nil {
		opt.KeyValueFunc = func(key, value string) string {
			return fmt.Sprintf("%s=%s", key, value)
		}
	}

	if opt.JoinSep == "" {
		opt.JoinSep = "&"
	}

	// 默认MD5
	if opt.HashFunc == nil {
		opt.HashFunc = func(s string) string {
			hash := md5.New()
			hash.Write([]byte(s))

			return hex.EncodeToString(hash.Sum(nil))
		}
	}

	var encoder = schema.NewEncoder()
	return func(i interface{}) (string, error) {
		values := url.Values{}
		if err := encoder.Encode(i, values); err != nil {
			return "", err
		}

		// url.Values
		var keys []string
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var pairsToSign []string
		for _, k := range keys {
			if contains(opt.IgnoreKeys, k) {
				continue
			}
			if opt.IgnoreEmpty && values.Get(k) == "" {
				continue
			}
			pairsToSign = append(pairsToSign, opt.KeyValueFunc(k, values.Get(k)))
		}

		strToSign := strings.Join(pairsToSign, opt.JoinSep)

		if opt.PostSignStrHook != nil {
			strToSign = opt.PostSignStrHook(strToSign)
		}

		return opt.HashFunc(strToSign), nil
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
