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

    const histories = db.collection("histories");

    const indexs = await histories.listIndexes().toArray();

    for (let j = 0; j < indexs.length; j++) {
      const element = indexs[j];
      if (element.name !== "_id_") {
        // 删除索引
        await histories.dropIndex(element.name);
      }
    }

    // 创建新的索引
    await histories.createIndex([
      "datastore_id",
      "created_at",
      "item_id",
      "history_type",
    ]);
    acount++;
  }

  change["処理概要-アプリまとめ"] = "Index更新数：" + acount;
  change["処理概要-終"] = "処理が正常に終了しました";

  return change;
}
