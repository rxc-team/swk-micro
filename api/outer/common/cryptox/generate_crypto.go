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
	"encoding/hex"
	"strings"
)

// GenerateMd5Password 生成MD5加密后的
func GenerateMd5Password(password, mail string) string {
	h := md5.New()
	h.Write([]byte(password))
	index := strings.Index(mail, "@")
	domain := mail[index+1:]
	return hex.EncodeToString(h.Sum([]byte(domain)))
}
