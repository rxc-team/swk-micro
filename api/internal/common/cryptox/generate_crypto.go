/*
 * @Description: 加解密的随机数生成器
 * @Author: RXC 廖云江
 * @Date: 2019-09-10 09:11:34
 * @LastEditors: Rxc 陳平
 * @LastEditTime: 2020-08-17 10:32:42
 */

package cryptox

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// GenerateMailAddress 生成邮箱地址
func GenerateMailAddress(mailName, domain string) string {
	var mailAddr strings.Builder
	// 添加名称
	mailAddr.WriteString(mailName)
	// 添加@
	mailAddr.WriteString("@")
	// 添加后缀名
	mailAddr.WriteString(domain)
	return mailAddr.String()
}

// GenerateMd5Password 生成MD5加密后的
func GenerateMd5Password(password, mail string) string {
	h := md5.New()
	h.Write([]byte(password))
	index := strings.Index(mail, "@")
	domain := mail[index+1:]
	return hex.EncodeToString(h.Sum([]byte(domain)))
}

// GenerateRandPassword 重置生成随机密码
func GenerateRandPassword() string {
	patternStr := "As%9"
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b) + patternStr
}
