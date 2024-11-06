async function run() {
    // Connect the client to the server
    await client.connect();
    // 获取系统所有客户情报
    // 日志情报
    const change = {};
    console.log("开始")
    change["処理概要開-始"] = "許可操作多言語対応";

    const db = client.db(`${data.db}_system`);
    const actionc = db.collection("actions");
    const actions = await actionc.find().toArray();
    for (let index = 0; index < actions.length; index++) {
        const action = actions[index];
        if (!action.action_name) {
            // 设置多语言数据
            const actionMap = {
                zh_CN: action.action_name_zh,
                en_US: action.action_name_en,
                ja_JP: action.action_name_ja,
                th_TH: "",
            }
            await actionc.updateOne({ _id: action._id }, { $set: { action_name: actionMap } }, { upsert: true })
        }
        // 去除无效数据
        await actionc.updateOne({ _id: action._id }, { $unset: { action_name_zh: null } })
        await actionc.updateOne({ _id: action._id }, { $unset: { action_name_en: null } })
        await actionc.updateOne({ _id: action._id }, { $unset: { action_name_ja: null } })
    }

    console.log("结束")
    change["処理概要-終"] = "処理が正常に終了しました";
    return change
}

