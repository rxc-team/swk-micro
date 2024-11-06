async function run() {
    // 日志情报
    var change = {};

    // 连接数据库
    await client.connect();

    // 获取系统所有客户情报
    var pit = client.db(data.db);
    var cc = pit.collection("customers");
    var customers = await cc.find().toArray();

    /* 顾客为空的场合直接返回 */
    if (!customers || customers.length === 0) {
        change["処理概要開-始"] = "顧客が存在しない、処理しない";
        return change;
    }

    // 処理対象
    change["処理概要-始"] =
        "処理対象顧客が" + customers.length + "個存在し、処理開始：";

    var acount = 0;
    /* 循环所有顾客，更新数据 */
    for (let i = 0; i < customers.length; i++) {
        // 单个顾客情报
        var cs = customers[i];
        var db = client.db(`${data.db}_${cs.customer_id}`);

        /* 查找顾客下的所有app情报 */
        var appc = db.collection("apps");
        var apps = await appc.find().toArray();

        /* app情报为空的场合，不需要改变 */
        if (!apps || apps.length === 0) {
            change["処理概要-アプリ"] = "アプリなしため、処理がスキップしました";
            continue;
        }
        /* 循环所有app情报，更新数据 */
        for (let k = 0; k < apps.length; k++) {
            const app = apps[k];
            acount++;

            /* 查找该app的所有config */
            const configc = db.collection("configs");
            const configs = await configc.find({ app_id: app.app_id }).toArray();
            /* config情报为空的场合，不需要改变 */
            if (!configs || configs.length === 0) {
                continue;
            }
            var appConfig = {}
            // 查找当前台账的满足条件的字段id
            for (let j = 0; j < configs.length; j++) {
                const config = configs[j]
                appConfig[`configs.${config.key}`] = config.value
            }

            await appc.updateOne(
                { app_id: app.app_id },
                { $set: appConfig }
            );
            // 删除当前app的原有的config信息。
            await configc.deleteMany({ app_id: app.app_id })
        }
        //判断configs集合是否存在，存在则删除
        const collections = await db.listCollections({ name: "configs" }).toArray();
        if (collections.length > 0) {
            await db.dropCollection("configs");
        }
    }

    change["処理概要-アプリまとめ"] = "アプリ更新数：" + acount;
    change["処理概要-終"] = "処理が正常に終了しました";

    return change;
}
