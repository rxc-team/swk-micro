async function run() {
    // Connect the client to the server
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

    let acount = 0;
    /* 循环所有顾客，更新数据 */
    for (let i = 0; i < customers.length; i++) {
        // 单个顾客情报
        const cs = customers[i];
        const db = client.db(`${data.db}_${cs.customer_id}`);
        console.log("处理顾客：" + cs.customer_id)
        // 查找顾客的所有语言数据
        const langc = db.collection("languages");
        const langueges = await langc.find().toArray();
        /* language情报为空的场合，不需要改变 */
        if (!langueges || langueges.length === 0) {
            change["処理概要-言語"] = "言語なしため、処理がスキップしました";
            continue;
        }
        const isAdd = langueges.find(language => language.lang_cd === "th-TH")
        if (isAdd) {
            continue;
        }
        // 设置泰语多语言数据
        let lang = {
            "domain": langueges[0].domain,
            "lang_cd": "th-TH",
            "text": "泰语",
            "abbr": "🇹🇭",
            "apps": {},
            "created_at": langueges[0].created_at,
            "created_by": langueges[0].created_by,
            "updated_at": langueges[0].updated_at,
            "updated_by": langueges[0].updated_by,
            "deleted_at": langueges[0].deleted_at,
            "deleted_by": langueges[0].deleted_by
        }
        lang["common.groups"] = {}
        // 插入泰语数据
        await langc.insertOne(lang)
        acount++;
    }
    console.log("结束")
    change["処理概要-処理対象顧客が"] = "処理対象顧客更新数：" + acount;
    change["処理概要-終"] = "処理が正常に終了しました";
    return change
}

