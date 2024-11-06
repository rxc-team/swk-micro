package mapping

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/import/common/floatx"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/common/stringx"
	"rxcsoft.cn/pit3/srv/import/model"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

const DefaultEmptyStr = "#N/A"

type OptionMap map[string]string

type bvParam struct {
	Index       int
	AppID       string
	Datastore   string
	DB          string
	Special     string
	Data        []string
	Header      []string
	UserList    []*user.User
	Fields      []*field.Field
	Relations   []*datastore.RelationItem
	OptionMap   map[string]OptionMap
	LangData    language.App
	MappingInfo datastore.MappingConf
	EmptyChange bool
}

// 编辑验证数据
func buildAndValidate(p bvParam) (result item.ChangeData, errorList []string) {

	// 判断数据项目数是否越界(超过header)
	if len(p.Data) > len(p.Header) {
		errorList = append(errorList, "incorrect number of row data items")
		return
	}

	rowItems := make(map[string]string)
	for index, it := range p.Data {
		rowItems[p.Header[index]] = it
	}

	query := make(map[string]*item.Value)
	change := make(map[string]*item.Value)
Loop:
	for _, mp := range p.MappingInfo.MappingRule {

		if col, ok := rowItems[mp.ToKey]; ok {
			// upsert insert update
			// 如果该字段作为主键，直接更新成空白即可
			if mp.PrimaryKey {
				// 如果当前值，等于默认空白更新内容，则直接替换为空白，然后继续进行后续判断
				if col == DefaultEmptyStr {
					rowItems[mp.ToKey] = ""
				}
			} else {
				// 如果是更新的场合
				// 如果空白表示不更新的情况下
				if p.MappingInfo.MappingType == "update" && !p.EmptyChange && len(col) == 0 {
					continue
				}
				// 如果当前值，等于默认空白更新内容，则直接替换为空白，然后继续进行后续判断
				if col == DefaultEmptyStr {
					rowItems[mp.ToKey] = ""
				}
			}
		}

		switch mp.DataType {
		case "text", "textarea":
			if mp.PrimaryKey {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断值是否为空
					if value != "" {
						// 判断是否包含无效特殊字符
						if stringx.SpecialCheck(value, p.Special) {
							query[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    value,
							}
						} else {
							// 该字段包含无效特殊字符
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] 無効な特殊文字があります。", p.Index, mp.ToKey))
							break Loop
						}

					} else {
						// 主键不能为空
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
						break Loop
					}
				} else {
					// 主键不能为空
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
					break Loop
				}
			} else {
				// 判断tokey是否有值，没有值直接使用默认值
				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
					change[mp.FromKey] = &item.Value{
						DataType: mp.DataType,
						Value:    mp.DefaultValue,
					}
				} else {
					// 判断是否导入该字段的数据
					if value, ok := rowItems[mp.ToKey]; ok {
						// 判断是否需要必须check
						if mp.IsRequired {
							// 判断值是否为空
							if value != "" {
								// 判断是否包含无效特殊字符
								if stringx.SpecialCheck(value, p.Special) {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								} else {
									// 该字段包含无效特殊字符
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] 無効な特殊文字があります。", p.Index, mp.ToKey))
									break Loop
								}
							} else {
								// 该字段是必须入力项目，一定要输入值
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
								break Loop
							}
						} else {
							// 判断值是否为空
							if value != "" {
								// 判断是否包含无效特殊字符
								if stringx.SpecialCheck(value, p.Special) {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								} else {
									// 该字段包含无效特殊字符
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] 無効な特殊文字があります。", p.Index, mp.ToKey))
									break Loop
								}
							} else {
								if len(mp.DefaultValue) > 0 {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    mp.DefaultValue,
									}
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								}
							}
						}
					} else {
						// 判断是否需要必须check
						if mp.IsRequired && len(mp.DefaultValue) == 0 {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						} else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    mp.DefaultValue,
							}
						}
					}
				}
			}
		case "number":
			if mp.PrimaryKey {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断值是否为空
					if value != "" {
						// 判断是否是数字类型
						_, e := strconv.ParseFloat(value, 64)
						if e != nil {
							fmt.Printf("number format has error : %v", e)
							// 不是正确的数字类型的数据。
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", p.Index, mp.ToKey))
							break Loop
						} else {
							query[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    value,
							}
						}
					} else {
						// 主键不能为空
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
						break Loop
					}
				} else {
					// 主键不能为空
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
					break Loop
				}
			} else {
				// 判断tokey是否有值，没有值直接使用默认值
				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
					// 判断是否是数字类型
					_, e := strconv.ParseInt(mp.DefaultValue, 10, 64)
					if e != nil {
						fmt.Printf("number format has error : %v", e)
						// 不是正确的数字类型的数据。
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", p.Index, mp.ToKey))
						break Loop
					} else {
						change[mp.FromKey] = &item.Value{
							DataType: mp.DataType,
							Value:    mp.DefaultValue,
						}
					}
				} else {
					// 判断是否导入该字段的数据
					if value, ok := rowItems[mp.ToKey]; ok {
						// 判断是否需要必须check
						if mp.IsRequired {
							// 判断值是否为空
							if value != "" {
								// 判断是否是数字类型
								nv, e := strconv.ParseFloat(value, 64)
								if e != nil {
									fmt.Printf("number format has error : %v", e)
									// 不是正确的数字类型的数据。
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", p.Index, mp.ToKey))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    floatx.ToFixedString(nv, getPrecision(p.Fields, mp.GetFromKey())),
									}
								}
							} else {
								// 该字段是必须入力项目，一定要输入值
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
								break Loop
							}
						} else {
							// 判断值是否为空
							if value != "" {
								// 判断是否是数字类型
								nv, e := strconv.ParseFloat(value, 64)
								if e != nil {
									fmt.Printf("number format has error : %v", e)
									// 不是正确的数字类型的数据。
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", p.Index, mp.ToKey))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    floatx.ToFixedString(nv, getPrecision(p.Fields, mp.GetFromKey())),
									}
								}
							} else {
								if len(mp.DefaultValue) > 0 {
									// 判断是否是数字类型
									_, e := strconv.ParseInt(mp.DefaultValue, 10, 64)
									if e != nil {
										fmt.Printf("number format has error : %v", e)
										// 不是正确的数字类型的数据。
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    mp.DefaultValue,
										}
									}
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								}
							}
						}
					} else {
						// 判断是否需要必须check
						if mp.IsRequired && len(mp.DefaultValue) == 0 {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						} else {
							// 判断是否是数字类型
							_, e := strconv.ParseInt(mp.DefaultValue, 10, 64)
							if e != nil {
								fmt.Printf("number format has error : %v", e)
								// 不是正确的数字类型的数据。
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", p.Index, mp.ToKey))
								break Loop
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    mp.DefaultValue,
								}
							}
						}
					}
				}
			}
		case "time":
			if mp.PrimaryKey {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断值是否为空
					if value != "" {
						// 判断是否是时间字符串
						_, e := time.Parse("15:04:05", value)
						if e != nil {
							fmt.Printf("time format has error : %v", e)
							// 不是正确的时间类型的数据。
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", p.Index, mp.ToKey))
							break Loop
						} else {
							query[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    value,
							}
						}
					} else {
						// 主键不能为空
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
						break Loop
					}
				} else {
					// 主键不能为空
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
					break Loop
				}
			} else {
				// 判断tokey是否有值，没有值直接使用默认值
				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
					// 判断是否是时间字符串
					_, e := time.Parse("15:04:05", mp.DefaultValue)
					if e != nil {
						fmt.Printf("time format has error : %v", e)
						// 不是正确的时间类型的数据。
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", p.Index, mp.ToKey))
						break Loop
					} else {
						change[mp.FromKey] = &item.Value{
							DataType: mp.DataType,
							Value:    mp.DefaultValue,
						}
					}
				} else {
					// 判断是否导入该字段的数据
					if value, ok := rowItems[mp.ToKey]; ok {
						// 判断是否需要必须check
						if mp.IsRequired {
							// 判断值是否为空
							if value != "" {
								// 判断是否是时间字符串
								_, e := time.Parse("15:04:05", value)
								if e != nil {
									fmt.Printf("time format has error : %v", e)
									// 不是正确的时间类型的数据。
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", p.Index, mp.ToKey))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								}
							} else {
								// 该字段是必须入力项目，一定要输入值
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
								break Loop
							}
						} else {
							// 判断值是否为空
							if value != "" {
								// 判断是否是时间字符串
								_, e := time.Parse("15:04:05", value)
								if e != nil {
									fmt.Printf("time format has error : %v", e)
									// 不是正确的时间类型的数据。
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", p.Index, mp.ToKey))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								}
							} else {
								if len(mp.DefaultValue) > 0 {
									// 判断是否是时间字符串
									_, e := time.Parse("15:04:05", mp.DefaultValue)
									if e != nil {
										fmt.Printf("time format has error : %v", e)
										// 不是正确的时间类型的数据。
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    mp.DefaultValue,
										}
									}
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								}
							}
						}
					} else {
						// 判断是否需要必须check
						if mp.IsRequired && len(mp.DefaultValue) == 0 {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						} else {
							// 判断是否是时间字符串
							_, e := time.Parse("15:04:05", mp.DefaultValue)
							if e != nil {
								fmt.Printf("time format has error : %v", e)
								// 不是正确的时间类型的数据。
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", p.Index, mp.ToKey))
								break Loop
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    mp.DefaultValue,
								}
							}
						}
					}
				}
			}
		case "date":
			if mp.PrimaryKey {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断值是否为空
					if value != "" {
						// 判断是否需要格式化日期
						if mp.Format != "" {
							ti, e := time.Parse(mp.Format, value)
							if e != nil {
								fmt.Printf("date format has error : %v", e)
								// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
								break Loop
							} else {
								query[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    ti.Format("2006-01-02"),
								}
							}
						} else {
							ti, e := time.Parse("2006-01-02", value)
							if e != nil {
								fmt.Printf("date format has error : %v", e)
								// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
								break Loop
							} else {
								query[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    ti.Format("2006-01-02"),
								}
							}
						}
					} else {
						// 主键不能为空
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
						break Loop
					}
				} else {
					// 主键不能为空
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", p.Index, mp.ToKey))
					break Loop
				}
			} else {
				// 判断tokey是否有值，没有值直接使用默认值
				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
					ti, e := time.Parse("2006-01-02", mp.DefaultValue)
					if e != nil {
						fmt.Printf("date format has error : %v", e)
						// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
						break Loop
					} else {
						change[mp.FromKey] = &item.Value{
							DataType: mp.DataType,
							Value:    ti.Format("2006-01-02"),
						}
					}
				} else {
					// 判断是否导入该字段的数据
					if value, ok := rowItems[mp.ToKey]; ok {
						// 判断是否需要必须check
						if mp.IsRequired {
							// 判断值是否为空
							if value != "" {
								// 判断是否需要格式化日期
								if mp.Format != "" {
									ti, e := time.Parse(mp.Format, value)
									if e != nil {
										fmt.Printf("date format has error : %v", e)
										// 日期格式化出错
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    ti.Format("2006-01-02"),
										}
									}
								} else {
									ti, e := time.Parse("2006-01-02", value)
									if e != nil {
										fmt.Printf("date format has error : %v", e)
										// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    ti.Format("2006-01-02"),
										}
									}
								}
							} else {
								// 该字段是必须入力项目，一定要输入值
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
								break Loop
							}
						} else {
							if value != "" {
								// 判断是否需要格式化日期
								if mp.Format != "" {
									ti, e := time.Parse(mp.Format, value)
									if e != nil {
										fmt.Printf("date format has error : %v", e)
										// 日期格式化出错
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    ti.Format("2006-01-02"),
										}
									}
								} else {
									ti, e := time.Parse("2006-01-02", value)
									if e != nil {
										fmt.Printf("date format has error : %v", e)
										// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    ti.Format("2006-01-02"),
										}
									}
								}
							} else {
								if len(mp.DefaultValue) > 0 {
									ti, e := time.Parse("2006-01-02", mp.DefaultValue)
									if e != nil {
										fmt.Printf("date format has error : %v", e)
										// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    ti.Format("2006-01-02"),
										}
									}
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    value,
									}
								}
							}
						}
					} else {
						// 判断是否需要必须check
						if mp.IsRequired && len(mp.DefaultValue) == 0 {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						} else {
							ti, e := time.Parse("2006-01-02", mp.DefaultValue)
							if e != nil {
								fmt.Printf("date format has error : %v", e)
								// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", p.Index, mp.ToKey))
								break Loop
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    ti.Format("2006-01-02"),
								}
							}
						}
					}
				}
			}
		case "switch":
			// 判断tokey是否有值，没有值直接使用默认值
			if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
				_, err := strconv.ParseBool(mp.DefaultValue)
				if err != nil {
					// 存在check错误
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
					break Loop
				} else {
					change[mp.FromKey] = &item.Value{
						DataType: mp.DataType,
						Value:    mp.DefaultValue,
					}
				}
			} else {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断是否需要必须check
					if mp.IsRequired {
						// 判断值是否为空
						if value != "" {
							_, err := strconv.ParseBool(value)
							if err != nil {
								// 存在check错误
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", p.Index, mp.ToKey, value))
								break Loop
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    value,
								}
							}
						} else {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						}
					} else {
						if len(mp.DefaultValue) > 0 {
							_, err := strconv.ParseBool(mp.DefaultValue)
							if err != nil {
								// 存在check错误
								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
								break Loop
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    mp.DefaultValue,
								}
							}
						} else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    value,
							}
						}
					}
				} else {
					// 判断是否需要必须check
					if mp.IsRequired && len(mp.DefaultValue) == 0 {
						// 该字段是必须入力项目，一定要输入值
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
						break Loop
					} else {
						_, err := strconv.ParseBool(mp.DefaultValue)
						if err != nil {
							// 存在check错误
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
							break Loop
						} else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    mp.DefaultValue,
							}
						}
					}
				}
			}
		case "user":
			// 判断tokey是否有值，没有值直接使用默认值
			if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
				uid := model.ReTranUser(mp.DefaultValue, p.UserList)
				if uid == "" {
					// 存在check错误
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
					break Loop
				} else {
					change[mp.FromKey] = &item.Value{
						DataType: mp.DataType,
						Value:    uid,
					}
				}
			} else {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断是否需要必须check
					if mp.IsRequired {
						// 判断值是否为空
						if value != "" {
							// 判断是单，还是多个用户
							inx := strings.LastIndex(value, ",")
							if inx == -1 {
								uid := model.ReTranUser(value, p.UserList)
								// 判断是否需要进行存在check
								if mp.Exist {
									if uid == "" {
										// 存在check错误
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, value))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    uid,
										}
									}
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    uid,
									}
								}
							} else {
								users := strings.Split(value, ",")
								var uids []string
								for _, u := range users {
									uid := model.ReTranUser(u, p.UserList)
									// 判断是否需要进行存在check
									if mp.Exist {
										if uid == "" {
											// 存在check错误
											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, u))
											break Loop
										} else {
											uids = append(uids, uid)
										}
									} else {
										uids = append(uids, uid)
									}
								}
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    strings.Join(uids, ","),
								}
							}
						} else {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						}
					} else {
						// 判断值是否为空
						if value != "" {
							// 判断是单，还是多个用户
							inx := strings.LastIndex(value, ",")
							if inx == -1 {
								uid := model.ReTranUser(value, p.UserList)
								// 判断是否需要进行存在check
								if mp.Exist {
									if uid == "" {
										// 存在check错误
										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, value))
										break Loop
									} else {
										change[mp.FromKey] = &item.Value{
											DataType: mp.DataType,
											Value:    uid,
										}
									}
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    uid,
									}
								}
							} else {
								users := strings.Split(value, ",")
								var uids []string
								for _, u := range users {
									uid := model.ReTranUser(u, p.UserList)
									// 判断是否需要进行存在check
									if mp.Exist {
										if uid == "" {
											// 存在check错误
											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, u))
											break Loop
										} else {
											uids = append(uids, uid)
										}
									} else {
										uids = append(uids, uid)
									}
								}
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    strings.Join(uids, ","),
								}
							}
						} else {
							if len(mp.DefaultValue) > 0 {
								uid := model.ReTranUser(mp.DefaultValue, p.UserList)
								if uid == "" {
									// 存在check错误
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    uid,
									}
								}
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    value,
								}
							}
						}
					}
				} else {
					// 判断是否需要必须check
					if mp.IsRequired && len(mp.DefaultValue) == 0 {
						// 该字段是必须入力项目，一定要输入值
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
						break Loop
					} else {
						uid := model.ReTranUser(mp.DefaultValue, p.UserList)
						if uid == "" {
							// 存在check错误
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
							break Loop
						} else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    uid,
							}
						}
					}
				}
			}
		case "options":
			// 判断tokey是否有值，没有值直接使用默认值
			if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
				oid := p.OptionMap[mp.FromKey][mp.DefaultValue]
				if oid == "" {
					// 存在check错误
					errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
					break Loop
				} else {
					change[mp.FromKey] = &item.Value{
						DataType: mp.DataType,
						Value:    oid,
					}
				}
			} else {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断是否需要必须check
					if mp.IsRequired {
						// 判断值是否为空
						if value != "" {
							// 获取选项ID
							oid := p.OptionMap[mp.FromKey][value]
							// 判断是否需要进行存在check
							if mp.Exist {
								if oid == "" {
									// 存在check错误
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", p.Index, mp.ToKey, value))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    oid,
									}
								}
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    oid,
								}
							}
						} else {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						}
					} else {
						// 判断值是否为空
						if value != "" {
							// 获取选项ID
							oid := p.OptionMap[mp.FromKey][value]
							// 判断是否需要进行存在check
							if mp.Exist {
								if oid == "" {
									// 存在check错误
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", p.Index, mp.ToKey, value))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    oid,
									}
								}
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    oid,
								}
							}
						} else {
							if len(mp.DefaultValue) > 0 {
								oid := p.OptionMap[mp.FromKey][mp.DefaultValue]
								if oid == "" {
									// 存在check错误
									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
									break Loop
								} else {
									change[mp.FromKey] = &item.Value{
										DataType: mp.DataType,
										Value:    oid,
									}
								}
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    value,
								}
							}
						}
					}
				} else {
					// 判断是否需要必须check
					if mp.IsRequired && len(mp.DefaultValue) == 0 {
						// 该字段是必须入力项目，一定要输入值
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
						break Loop
					} else {
						oid := p.OptionMap[mp.FromKey][mp.DefaultValue]
						if oid == "" {
							// 存在check错误
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
							break Loop
						} else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    oid,
							}
						}
					}
				}
			}
		case "lookup":
			// 判断tokey是否有值，没有值直接使用默认值
			if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
				// lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, mp.DefaultValue)
				// if lv == "" {
				// 	// 存在check错误
				// 	errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
				// 	break Loop
				// } else {
				change[mp.FromKey] = &item.Value{
					DataType: mp.DataType,
					Value:    mp.DefaultValue,
				}
				// }
			} else {
				// 判断是否导入该字段的数据
				if value, ok := rowItems[mp.ToKey]; ok {
					// 判断是否需要必须check
					if mp.IsRequired {
						// 判断值是否为空
						if value != "" {
							// 获取关联数据
							// lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, value)
							// // 判断是否需要进行存在check
							// if mp.Exist {
							// 	if lv == "" {
							// 		// 存在check错误
							// 		errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", p.Index, mp.ToKey, value))
							// 		break Loop
							// 	} else {
							// 		change[mp.FromKey] = &item.Value{
							// 			DataType: mp.DataType,
							// 			Value:    lv,
							// 		}
							// 	}
							// } else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    value,
							}
							// }
						} else {
							// 该字段是必须入力项目，一定要输入值
							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
							break Loop
						}
					} else {
						// 判断值是否为空
						if value != "" {
							// // 获取关联数据
							// lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, value)
							// // 判断是否需要进行存在check
							// if mp.Exist {
							// 	if lv == "" {
							// 		// 存在check错误
							// 		errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", p.Index, mp.ToKey, value))
							// 		break Loop
							// 	} else {
							// 		change[mp.FromKey] = &item.Value{
							// 			DataType: mp.DataType,
							// 			Value:    lv,
							// 		}
							// 	}
							// } else {
							change[mp.FromKey] = &item.Value{
								DataType: mp.DataType,
								Value:    value,
							}
							// }
						} else {
							if len(mp.DefaultValue) > 0 {
								// lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, mp.DefaultValue)
								// if lv == "" {
								// 	// 存在check错误
								// 	errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
								// 	break Loop
								// } else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    mp.DefaultValue,
								}
								// }
							} else {
								change[mp.FromKey] = &item.Value{
									DataType: mp.DataType,
									Value:    "",
								}
							}

						}
					}
				} else {
					// 判断是否需要必须check
					if mp.IsRequired && len(mp.DefaultValue) == 0 {
						// 该字段是必须入力项目，一定要输入值
						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", p.Index, mp.ToKey))
						break Loop
					} else {
						// lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, mp.DefaultValue)
						// if lv == "" {
						// 	// 存在check错误
						// 	errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", p.Index, mp.ToKey, mp.DefaultValue))
						// 	break Loop
						// } else {
						change[mp.FromKey] = &item.Value{
							DataType: mp.DataType,
							Value:    mp.DefaultValue,
						}
						// }
					}
				}
			}
		}
	}

	itemService := item.NewItemService("database", client.DefaultClient)

	// 检查关系是否存在
	for _, rat := range p.Relations {

		var keys []string

		param := item.CountRequest{
			AppId:         p.AppID,
			DatastoreId:   rat.DatastoreId,
			ConditionList: []*item.Condition{},
			ConditionType: "and",
			Database:      p.DB,
		}

		// 第一步，判断当前关系字段是否存在于传入数据中
		var existCount = 0
		var emptyCount = 0
		for relationKey, localKey := range rat.Fields {
			name := p.LangData.Fields[p.Datastore+"_"+localKey]
			keys = append(keys, name)

			if val, ok := query[localKey]; ok {
				if len(val.Value) == 0 {
					emptyCount++
				}

				param.ConditionList = append(param.ConditionList, &item.Condition{
					FieldId:     relationKey,
					FieldType:   "text",
					SearchValue: val.GetValue(),
					Operator:    "=",
					IsDynamic:   true,
				})
				existCount++
			}
			if val, ok := change[localKey]; ok {
				if len(val.Value) == 0 {
					emptyCount++
				}

				param.ConditionList = append(param.ConditionList, &item.Condition{
					FieldId:     relationKey,
					FieldType:   "text",
					SearchValue: val.GetValue(),
					Operator:    "=",
					IsDynamic:   true,
				})
				existCount++
			}
		}

		// 如果全部不存在,直接跳过
		if existCount == 0 {
			continue
		}

		// 如果部分存在
		if existCount < len(rat.Fields) {
			errorList = append(errorList, fmt.Sprintf("%v関連アイテムには、対応するすべてのデータが渡されていません", keys))
			continue
		}

		// 如果当前关系的所有值都是空，则跳过检查
		if emptyCount == len(rat.Fields) {
			continue
		}

		response, err := itemService.FindCount(context.TODO(), &param)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("%v関連アイテムのデータは存在しません。", keys))
			continue
		}

		if response.GetTotal() == 0 {
			errorList = append(errorList, fmt.Sprintf("%v関連アイテムのデータは存在しません。", keys))
			continue
		}
	}

	result = item.ChangeData{
		Query:  query,
		Change: change,
		Index:  int64(p.Index),
	}

	return
}

// 获取mapping信息
func getMappingInfo(db, datastoreID, mappingID string) (*datastore.MappingConf, error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.MappingRequest
	req.DatastoreId = datastoreID
	req.MappingId = mappingID
	req.Database = db

	response, err := datastoreService.FindDatastoreMapping(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getMappingInfo", err.Error())
		return nil, err
	}

	return response.GetMapping(), nil
}

// 获取精度
func getPrecision(fs []*field.Field, fid string) int64 {
	for _, f := range fs {
		if f.FieldId == fid {
			return f.Precision
		}
	}

	return 0
}

// 获取台账信息
func getDatastore(db, ds string) (*datastore.Datastore, error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreRequest
	// 从path获取
	req.DatastoreId = ds
	req.Database = db
	response, err := datastoreService.FindDatastore(context.TODO(), &req)
	if err != nil {
		return nil, err
	}
	return response.GetDatastore(), nil
}
