async function run() {
  // 日志情报
  var change = {};

  // 连接数据库
  await client.connect();

  // 获取系统所有客户情报
  var pit = client.db("pit");
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

  var dcount = 0;
  /* 循环所有顾客，更新数据 */
  for (let i = 0; i < customers.length; i++) {
    // 单个顾客情报;
    var cs = customers[i];
    var db = client.db(`pit_${cs.customer_id}`);

    /* 查找顾客下的所有台账 */
    var dc = db.collection("data_stores");
    var ds = await dc.find().toArray();

    /* 台账为空的场合，不需要改变 */
    if (!ds || ds.length === 0) {
      change["処理概要-台帳"] = "台帳なしため、処理がスキップしました";
      continue;
    }
    /* 循环所有台账，更新数据 */
    for (let j = 0; j < ds.length; j++) {
      var datastore = ds[j];
      dcount++;

      /* 查找该台账的所有字段 */
      const fsc = db.collection("fields");
      const fields = await fsc
        .find({ datastore_id: datastore.datastore_id })
        .toArray();

      // 查找当前台账的满足条件的字段id
      for (let m = 0; m < fields.length; m++) {
        const field = fields[m];
        if (field.unique === true || field.field_type === "autonum") {
          await dc.updateOne(
            { datastore_id: datastore.datastore_id },
            { $addToSet: { unique_fields: field.field_id } }
          );
        }
      }
    }
  }

  change["処理概要-台帳まとめ"] = "台帳更新数：" + dcount;
  change["処理概要-終"] = "処理が正常に終了しました";

  return change;
}
