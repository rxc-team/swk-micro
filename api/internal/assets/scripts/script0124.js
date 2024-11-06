// 为添加履历映射下载机能,更新台账映射情报
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
  var fcount = 0;
  /* 循环所有顾客，更新数据 */
  for (let i = 0; i < customers.length; i++) {
    // 单个顾客情报
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
      // 台账
      var d = ds[j];
      /* 循环所有映射情报，更新映射应用类型情报 */
      for (let k = 0; k < d.mappings.length; k++) {
        var setMpData = {};
        // 映射
        var mp = d.mappings[k];
        if (!mp.hasOwnProperty("apply_type")) {
          dcount++;
          setMpData["mappings.$.apply_type"] = "datastore";
          // 更新映射应用类型开始
          change[
            "CustomerID[" +
              cs.customer_id +
              "]--DatastoreID[" +
              d.datastore_id +
              "]--MappingID[" +
              mp.mapping_id +
              "]--start"
          ] = "update-start";

          // 更新映射应用类型
          await dc.updateOne(
            { _id: d._id, "mappings.mapping_id": mp.mapping_id },
            { $set: setMpData }
          );

          // 更新映射应用类型结束
          change[
            "CustomerID[" +
              cs.customer_id +
              "]--DatastoreID[" +
              d.datastore_id +
              "]--MappingID[" +
              mp.mapping_id +
              "]--end"
          ] = "update-end";
        }

        /* 循环所有映射字段情报，更新映射字段检查变更情报 */
        for (let m = 0; m < mp.mapping_rule.length; m++) {
          var setData = {};
          // 映射字段
          var mpf = mp.mapping_rule[m];
          if (!mpf.hasOwnProperty("check_change")) {
            fcount++;
            setData[
              "mappings.$[outer].mapping_rule.$[inner].check_change"
            ] = false;

            var ukeyName =
              mpf.from_key === ""
                ? "ToKeyID[" + mpf.to_key
                : "FromKeyID[" + mpf.from_key;

            change[
              "CustomerID[" +
                cs.customer_id +
                "]--DatastoreID[" +
                d.datastore_id +
                "]--MappingID[" +
                mp.mapping_id +
                "]--" +
                ukeyName +
                "]--start"
            ] = "update-start";

            await dc.updateOne(
              { _id: d._id },
              { $set: setData },
              {
                arrayFilters: [
                  {
                    "outer.mapping_id": mp.mapping_id,
                  },
                  {
                    $or: [
                      {
                        "inner.from_key": mpf.from_key,
                      },
                      {
                        "inner.to_key": mpf.to_key,
                      },
                    ],
                  },
                ],
              }
            );

            change[
              "CustomerID[" +
                cs.customer_id +
                "]--DatastoreID[" +
                d.datastore_id +
                "]--MappingID[" +
                mp.mapping_id +
                "]--" +
                ukeyName +
                "]--end"
            ] = "update-end";
          }
        }
      }
    }
  }

  change["処理概要-台帳まとめ"] = "台帳更新数：" + dcount;
  change["処理概要-キーまとめ"] = "マッピングキー更新数：" + fcount;
  change["処理概要-終"] = "処理が正常に終了しました";

  return change;
}
