async function run() {
    // 日志情报
    const change = {};

    // 连接数据库
    await client.connect();

    // 获取系统所有客户情报
    const pit = client.db(data.db);
    const cc = pit.collection('customers');
    const customers = await cc.find().toArray();

    /* 顾客为空的场合直接返回 */
    if (!customers || customers.length === 0) {
        console.log('顧客が存在しない、処理しない');
        // change['処理概要開-始'] = '顧客が存在しない、処理しない';
        return change;
    }

    // 処理対象
    console.log('処理対象顧客が' + customers.length + '個存在し、処理開始：');
    change['処理概要-始'] = '処理対象顧客が' + customers.length + '個存在し、処理開始：';

    var dcount = 0;
    /* 循环所有顾客，更新数据 */
    for (let i = 0; i < customers.length; i++) {
        // 单个顾客情报
        const cs = customers[i];
        const db = client.db(`${data.db}_${cs.customer_id}`);

        // 获取顾客的所有用户
        const userList = await db.collection('users').find().toArray();

        // 获取语言数据
        const langData = await db.collection('languages').findOne({ lang_cd: 'ja-JP' });

        /* 查找顾客下的所有台账 */
        const dc = db.collection('data_stores');
        const ds = await dc.find().toArray();

        /* 台账为空的场合，不需要改变 */
        if (!ds || ds.length === 0) {
            console.log('没有台账数据');
            continue;
        }
        /* 循环所有台账，更新数据 */
        for (let j = 0; j < ds.length; j++) {
            datastore = ds[j];
            dcount++;
            console.log(`开始处理台账【${datastore.datastore_id}】`);

            await db.collection('data_histories').deleteMany({ "datastore_id": datastore.datastore_id });
            await db.collection('field_histories').deleteMany({ "datastore_id": datastore.datastore_id });

            /* 查找该台账的所有字段 */
            const fsc = db.collection('fields');
            const fields = await fsc.find({ datastore_id: datastore.datastore_id }).toArray();

            const hsc = db.collection('histories');
            const cursor = hsc.find({ datastore_id: datastore.datastore_id });

            const dataHisotryList = [];
            const fieldHisotryList = [];

            let index = 1;
            while (await cursor.hasNext()) {
                const hs = await cursor.next();
                const hid = new Date(hs.created_at).format('yyyyMMddhhmmssS') + i.toString() + j.toString() +index.toString();

                // 生成data履历数据
                const fixed_items = hs.fixed_items;
                // 循环，修改原本的内容（user，options）
                for (const key in fixed_items) {
                    if (Object.hasOwnProperty.call(fixed_items, key)) {
                        const item = fixed_items[key];

                        const fs = fields.find(f => f.field_id === key);
                        // 当前字段存在的场合
                        if (fs) {
                            // TODO()
                            if (item.data_type === 'user') {
                                if (item.value instanceof Array && item.value.length > 0) {
                                    const users = item.value;
                                    // user转换
                                    users.forEach(user => {
                                        let us = userList.find(u => user === u.user_id)
                                        if (us) {
                                            item.value.push(us.user_name)
                                        } else {
                                            let value = user + '(Delete)'
                                            item.value.push(value)
                                        }
                                    });
                                }
                            }
                            if (item.data_type === 'file') {
                                if (item.value && item.value instanceof Array) {
                                    item.value = JSON.stringify(item.value);
                                }
                                // file转换
                            }
                            if (item.data_type === 'lookup') {
                                if (item.value && typeof item.value === 'object') {
                                    item.value = item.value.value;
                                }
                                // lookup转换
                            }
                            if (item.data_type === 'options') {
                                // option转换
                                if (item.value) {
                                    if (Object.hasOwnProperty.call(langData.apps, fs.app_id)) {
                                        const v = langData.apps[fs.app_id].options[`${fs.option_id}_${item.value}`];
                                        if (v) {
                                            item.value = v;
                                        } else {
                                            item.value = item.value + '(Delete)';
                                        }
                                    } else {
                                        item.value = item.value + '(Delete)';
                                    }
                                }
                            }
                        }
                    }
                }

                var type = hs.history_type;
                if (type === 'new') {
                    type = 'insert';
                }

                const dataHs = {
                    history_id: hid,
                    history_type: type,
                    datastore_id: hs.datastore_id,
                    item_id: hs.item_id,
                    fixed_items: fixed_items,
                    created_at: hs.created_at,
                    created_by: hs.created_by
                };

                dataHisotryList.push(dataHs);

                // 生成字段履历数据
                for (const key in hs.items) {
                    if (Object.hasOwnProperty.call(hs.items, key)) {
                        const change = hs.items[key];

                        const fs = fields.find(f => f.field_id === key);
                        // 该字段存在的场合
                        if (fs) {
                            let o = change.old_value;
                            // 替换null结果
                            if (o === 'null') {
                                o = '';
                            }
                            let n = change.new_value;
                            // 替换null结果
                            if (n === 'null') {
                                n = '';
                            }

                            if (change.data_type === 'file' && o !== '') {
                                const ofs = JSON.parse(o);
                                if (ofs) {
                                    o = ofs.map(f => f.name).join(',');
                                }
                                const nfs = JSON.parse(n);
                                if (nfs) {
                                    n = nfs.map(f => f.name).join(',');
                                }
                            }
                            // TODO()
                            if (change.data_type === 'user') {
                                const oldValue = []
                                const newValue = []
                                if (o) {
                                    const ofs = getUsers(o);
                                    ofs.forEach(user => {
                                        let us = userList.find(u => user === u.user_id)
                                        if (us) {
                                            oldValue.push(us.user_name)
                                        } else {
                                            let value = user + '(Delete)'
                                            oldValue.push(value)
                                        }
                                    });
                                    o = oldValue.join(",")
                                } else {
                                    o = ""
                                }
                                if (n) {
                                    const nfs = getUsers(n);
                                    nfs.forEach(user => {
                                        let us = userList.find(u => user === u.user_id)
                                        if (us) {
                                            newValue.push(us.user_name)
                                        } else {
                                            let value = user + '(Delete)'
                                            newValue.push(value)
                                        }
                                    });
                                    n = newValue.join(",")
                                } else {
                                    n = ""
                                }
                                console.log("old:%v", o)
                                console.log("new:%v", n)
                            }

                            if (change.data_type === 'options') {
                                if (Object.hasOwnProperty.call(langData.apps, fs.app_id)) {
                                    if (o != '') {
                                        const v = langData.apps[fs.app_id].options[`${fs.option_id}_${o}`];
                                        if (v) {
                                            o = v;
                                        } else {
                                            o = o + '(Delete)';
                                        }
                                    }
                                    if (n != '') {
                                        const v = langData.apps[fs.app_id].options[`${fs.option_id}_${n}`];
                                        if (v) {
                                            n = v;
                                        } else {
                                            n = n + '(Delete)';
                                        }
                                    }
                                }
                            }

                            let local_name = '';

                            if (Object.hasOwnProperty.call(langData.apps, fs.app_id)) {
                                const v = langData.apps[fs.app_id].fields[`${fs.datastore_id}_${key}`];
                                if (v) {
                                    local_name = v;
                                }
                            }

                            const fH = {
                                history_id: hid,
                                history_type: type,
                                datastore_id: hs.datastore_id,
                                item_id: hs.item_id,
                                field_id: key,
                                local_name: local_name,
                                field_name: change.field_name,
                                old_value: o,
                                new_value: n,
                                created_at: hs.created_at,
                                created_by: hs.created_by
                            };

                            fieldHisotryList.push(fH);
                        }
                    }
                }

                index++;
            }

            let dataHistores = []
            if (dataHisotryList.length > 0) {
                for (let i = 0; i < dataHisotryList.length; i++) {
                    dataHistores.push(dataHisotryList[i])
                    if (dataHistores.length === 500) {
                        await db.collection('data_histories').insertMany(dataHistores);
                        dataHistores = []
                    }
                }
                await db.collection('data_histories').insertMany(dataHistores);
            }
            let fieldHistores = []
            if (fieldHisotryList.length > 0) {
                for (let j = 0; j < fieldHisotryList.length; j++) {
                    fieldHistores.push(fieldHisotryList[j])
                    if (fieldHistores.length === 500) {
                        await db.collection('field_histories').insertMany(fieldHistores);
                        fieldHistores = []
                    }
                }
                await db.collection('field_histories').insertMany(fieldHistores);
            }

            console.log(`处理结束台账【${datastore.datastore_id}】`);
        }
    }

    console.log(`处理了共${dcount}】台账`);
    change['処理概要-台帳まとめ'] = '台帳更新数：' + dcount;
    change['処理概要-終'] = '処理が正常に終了しました';

    return change;
}
Date.prototype.format = function (fmt) {
    var o = {
        'M+': this.getMonth() + 1, //月份
        'd+': this.getDate(), //日
        'h+': this.getHours(), //小时
        'm+': this.getMinutes(), //分
        's+': this.getSeconds(), //秒
        'q+': Math.floor((this.getMonth() + 3) / 3), //季度
        S: this.getMilliseconds() //毫秒
    };
    if (/(y+)/.test(fmt)) {
        fmt = fmt.replace(RegExp.$1, (this.getFullYear() + '').substr(4 - RegExp.$1.length));
    }
    for (var k in o) {
        if (new RegExp('(' + k + ')').test(fmt)) {
            fmt = fmt.replace(RegExp.$1, RegExp.$1.length == 1 ? o[k] : ('00' + o[k]).substr(('' + o[k]).length));
        }
    }
    return fmt;
};

function getUsers(value) {
    let pattern = /[0-9a-z]+/
    return pattern.exec(value);
} 
