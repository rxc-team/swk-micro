async function run() {
    await client.connect();
    // 获取系统所有客户情报
    // 日志情报
    const change = {};
    const pit = client.db(data.db);
    const cc = pit.collection("customers");
    const customers = await cc.find().toArray();

    /* 顾客为空的场合直接返回 */
    if (!customers || customers.length === 0) {
        change["処理概要開-始"] = "顧客が存在しない、処理しない";
        return change;
    }

    // 処理対象
    change["処理概要-始"] =
        "処理対象顧客が" + customers.length + "個存在し、処理開始：";

    let approveCount = 0;
    const lang = "ja-JP";
    /* 循环所有顾客，更新数据 */
    for (let i = 0; i < customers.length; i++) {
        // 单个顾客情报
        const cs = customers[i];
        const db = client.db(`${data.db}_${cs.customer_id}`);


        /* 查找该顾客的所有users信息，保存到map中 */
        const userMap = new Map()
        const userc = db.collection("users");
        const users = await userc.find().toArray();
        for (let j = 0; j < users.length; j++) {
            const user = users[j];
            userMap.set(user.user_id, user.user_name)
        }


        /* 查找该顾客的当前语言信息 */
        const langc = db.collection("languages");
        const langguage = await langc.findOne({ lang_cd: lang })

        /* 查找该顾客的所有approves */
        const approvec = db.collection("approves");
        const approves = await approvec.find().toArray();
        for (let k = 0; k < approves.length; k++) {
            const approve = approves[k];

            // 查询台账下的选项字段的数据
            const fieldc = db.collection("fields")
            const opfs = await fieldc.find({ field_type: "options", app_id: approve.app_id, datastore_id: approve.datastore_id }).toArray();

            // 需要更新的approves履历
            const newApproves = {}
            const current = approve.current
            if (current) {
                continue;
            }
            // 更新history中的option和user的值
            const history = approve.history
            // 判断history是否为对象，和对象是否有键
            if (!(Object.prototype.isPrototypeOf(history) && Object.keys(history).length === 0)) {

                for (const key in history) {
                    // 判断当前对象中是否存在指定的key
                    if (Object.hasOwnProperty.call(history, key)) {
                        const fValue = history[key];
                        if (fValue.data_type == "options") {
                            const optField = opfs.find(opf => opf.field_id === key)
                            // 获取多语言optionlable数据
                            let optLable = ""
                            if (optField) {
                                if (Object.hasOwnProperty.call(langguage.apps, approve.app_id)) {
                                    optKey = optField.option_id + '_' + fValue.value
                                    if (Object.hasOwnProperty.call(langguage.apps[approve.app_id].options, optKey)) {
                                        optLable = langguage.apps[approve.app_id].options[optKey]
                                    } else {
                                        optLable = optKey + "(Delete)"
                                    }
                                } else {
                                    optLable = approve.app_id + "(Delete)"
                                }
                            } else {
                                optLable = key + "(Delete)"
                            }
                            newApproves["history." + key + ".value"] = optLable
                        }
                        if (fValue.data_type == "user") {
                            const userNames = []
                            for (let m = 0; m < fValue.value.length; m++) {
                                const userID = fValue.value[m];
                                // 设置user的名称
                                if (userMap.has(userID)) {
                                    userNames.push(userMap.get(userID))
                                } else {
                                    const userName = userID + "(Delete)"
                                    userNames.push(userName)
                                }
                            }
                            newApproves["history." + key + ".value"] = userNames
                        }
                    }
                }

            }
            // 添加current数据
            const items = approve.items
            newApproves["current"] = items
            if (!(Object.prototype.isPrototypeOf(items) && Object.keys(items).length === 0)) {
                for (const key in items) {
                    if (Object.hasOwnProperty.call(items, key)) {
                        const fValue = items[key];
                        if (fValue.data_type == "options") {
                            optField = opfs.find(opf => opf.field_id === key);
                            // 获取多语言optionlable数据
                            let optLable = ""
                            if (optField) {
                                if (Object.hasOwnProperty.call(langguage.apps, approve.app_id)) {
                                    optKey = optField.option_id + '_' + fValue.value
                                    if (Object.hasOwnProperty.call(langguage.apps[approve.app_id].options, optKey)) {
                                        optLable = langguage.apps[approve.app_id].options[optKey]
                                    } else {
                                        optLable = optKey + "(Delete)"
                                    }
                                } else {
                                    optLable = approve.app_id + "(Delete)"
                                }
                            } else {
                                optLable = key + "(Delete)"
                            }
                            newApproves.current[key].value = optLable
                        }
                        if (fValue.data_type == "user") {
                            var currentUserNames = []
                            for (let n = 0; n < fValue.value.length; n++) {
                                const userID = fValue.value[n];
                                // 设置user的名称
                                if (userMap.has(userID)) {
                                    currentUserNames.push(userMap.get(userID))
                                } else {
                                    const userName = userID + "(Delete)"
                                    currentUserNames.push(userName)
                                }
                            }
                            newApproves.current[key].value = currentUserNames
                        }
                    }
                }
            }
            await approvec.updateOne({ _id: approve._id }, { $set: newApproves })
            approveCount++;
        }
    }
    change["処理概要-承認まとめ"] = "承認更新数：" + approveCount;
    change["処理概要-終"] = "処理が正常に終了しました";

    return change;
}
