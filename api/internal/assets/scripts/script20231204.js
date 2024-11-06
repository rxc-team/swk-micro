async function run() {
    // Connect the client to the server
    await client.connect();

    // 日志情报
    const change = {};

    // 获取系统所有客户情报
    const pit = client.db("pit");
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
        const db = client.db(`pit_${cs.customer_id}`);
        console.log("处理顾客：" + cs.customer_id)

        /* 查找顾客下的所有台账 */
        const dc = db.collection("data_stores");
        const ds = await dc.find().toArray();

        /* 台账为空的场合，不需要改变 */
        if (!ds || ds.length === 0) {
            change["処理概要-台帳"] = "台帳なしため、処理がスキップしました";
            continue;
        }
        
        /* 循环所有台账，更新数据 */
        for (let j = 0; j < ds.length; j++) {
            const datastore = ds[j];
            const segmentcd = ["segmentcd"];
            const locationcd = ["locationcd"];
            if (datastore.api_key === "segment") {
                const se = db.collection(`item_${datastore.datastore_id}`);
                const indexs = await se.listIndexes().toArray();
                for (let i = 0; i < indexs.length; i++) {
                    const element = indexs[i];
                    if (element.name == "items.segmentnm.value_1") {
                        // 删除索引
                        await se.dropIndex(element.name);
                    }
                }
                acount++;
                await dc.updateOne(
                    { datastore_id: datastore.datastore_id },
                    {
                        $set: {
                            unique_fields: segmentcd,
                        }
                    }
                );
            }
            if (datastore.api_key === "location") {
                const lo = db.collection(`item_${datastore.datastore_id}`);
                const indexs = await lo.listIndexes().toArray();
                for (let i = 0; i < indexs.length; i++) {
                    const element = indexs[i];
                    if (element.name == "items.locationnm.value_1") {
                        // 删除索引
                        await lo.dropIndex(element.name);
                    }
                }
                acount++;
                await dc.updateOne(
                    { datastore_id: datastore.datastore_id },
                    {
                        $set: {
                            unique_fields: locationcd,
                        }
                    }
                );
            }
        }
    }
    console.log("结束")
    change["処理概要-データまとめ"] = "データ更新の総数：" + acount;
    change["処理概要-終"] = "処理が正常に終了しました";

    return change
}