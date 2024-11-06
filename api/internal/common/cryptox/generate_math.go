/*
 * @Descripttion: 伪随机数生成器
 * @Author: Rxc 陳平
 * @Date: 2020-08-06 16:49:59
 * @LastEditors: Rxc 陳平
 * @LastEditTime: 2020-08-06 17:19:45
 */

package cryptox

import (
	"math/rand"
	"strconv"
)

// GenerateRandCaptcha 生成随机验证码
func GenerateRandCaptcha() string {
	//随机生成6位数的验证码
	a := rand.Int() % 1000000
	return strconv.Itoa(a)
}
