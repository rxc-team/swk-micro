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
        // 判断数据表是否存在，不存在则创建
        const dataCursor = db.listCollections({ name: "data_histories" })
        const dataResult = await dataCursor.hasNext()
        if (!dataResult) {
            db.createCollection("data_histories")
        }
        const fieldCursor = db.listCollections({ name: "field_histories" })
        const fieldResult = await fieldCursor.hasNext()
        if (!fieldResult) {
            db.createCollection("field_histories")
        }

        const dataHistories = db.collection("data_histories")
        const fieldHistories = db.collection("field_histories")

        const dataIndexs = await dataHistories.listIndexes().toArray();

        for (let j = 0; j < dataIndexs.length; j++) {
            const element = dataIndexs[j];
            if (element.name !== "_id_") {
                // 删除索引
                await dataHistories.dropIndex(element.name);
            }
        }
        const fieldIndexs = await fieldHistories.listIndexes().toArray();

        for (let j = 0; j < fieldIndexs.length; j++) {
            const element = fieldIndexs[j];
            if (element.name !== "_id_") {
                // 删除索引
                await fieldHistories.dropIndex(element.name);
            }
        }

        // 创建新的索引
        await dataHistories.createIndex({ "history_id": 1 }, { unique: true });
        await dataHistories.createIndex([
            {
                "datastore_id": 1,
                "created_at": 1,
                "item_id": 1,
                "history_type": 1
            }
        ]);
        await fieldHistories.createIndex({ "history_id": 1 });
        await fieldHistories.createIndex([
            {
                "datastore_id": 1,
                "created_at": 1,
                "field_id": 1,
                "item_id": 1,
                "history_type": 1
            }
        ]);
        acount++;
    }
    console.log("结束")
    change["処理概要-処理対象顧客が"] = "処理対象顧客更新数：" + acount;
    change["処理概要-終"] = "処理が正常に終了しました";
    return change
}

