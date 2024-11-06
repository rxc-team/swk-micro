async function run() {
  await client.connect();

  const pit = client.db("pit");
  const cc = pit.collection("customers");

  const customers = await cc.find().toArray();

  const change = {};

  /* 循环所有顾客，更新数据 */
  for (let i = 0; i < customers.length; i++) {
    const cs = customers[i];

    const db = client.db(`pit_${cs.customer_id}`);
    /* 查找顾客下的所有台账 */
    const dsc = db.collection("data_stores");
    const datastores = await dsc.find().toArray();

    /* 循环所有台账，更新台账数据 */
    for (let j = 0; j < datastores.length; j++) {
      const ds = datastores[j];

      /* 查找该台账的所有字段 */
      const fsc = db.collection("fields");
      const fields = await fsc
        .find({ datastore_id: ds.datastore_id })
        .toArray();

      /* 如果没有自动採番字段，则跳过 */
      if (
        !fields ||
        fields.filter((f) => f.field_type === "autonum").length === 0
      ) {
        continue;
      }

      const ic = db.collection(`item_${ds.datastore_id}`);

      /* 编辑更新数据 */
      const pipe = [];

      /* 变更处理 */
      const project = {
        _id: 1,
        item_id: 1,
        app_id: 1,
        datastore_id: 1,
        owners: 1,
        check_type: 1,
        check_status: 1,
        created_at: 1,
        created_by: 1,
        updated_at: 1,
        updated_by: 1,
        checked_at: 1,
        checked_by: 1,
        label_time: 1,
        status: 1,
      };

      fields.forEach((f) => {
        if (f.field_type === "autonum") {
          var autonum = {
            $cond: {
              else: {
                $cond: {
                  else: {
                    $concat: [
                      {
                        $reduce: {
                          in: {
                            $concat: ["$$value", "0"],
                          },
                          initialValue: "pre_xxxx",
                          input: {
                            $range: [
                              0,
                              {
                                $subtract: [
                                  6,
                                  {
                                    $strLenCP: {
                                      $toString: {
                                        $ifNull: [
                                          "$items.field_xxxx.value",
                                          "0",
                                        ],
                                      },
                                    },
                                  },
                                ],
                              },
                            ],
                          },
                        },
                      },
                      {
                        $toString: {
                          $ifNull: ["$items.field_xxxx.value", "0"],
                        },
                      },
                    ],
                  },
                  if: {
                    $lt: [
                      {
                        $substr: [
                          {
                            $toString: {
                              $ifNull: ["$items.field_xxxx.value", "0"],
                            },
                          },
                          0,
                          -1,
                        ],
                      },
                      6,
                    ],
                  },
                  then: {
                    $concat: [
                      "d",
                      {
                        $toString: {
                          $ifNull: ["$items.field_xxxx.value", "0"],
                        },
                      },
                    ],
                  },
                },
              },
              if: {
                $eq: [
                  {
                    $type: "$items.field_xxxx.value",
                  },
                  "string",
                ],
              },
              then: "$items.field_xxxx.value",
            },
          };

          var afs = JSON.stringify(autonum);

          afs = afs.replace(/field_xxxx/g, f.field_id);
          afs = afs.replace(/pre_xxxx/g, f.prefix);

          var auto = JSON.parse(afs);

          project["items." + f.field_id + ".data_type"] = "autonum";
          project["items." + f.field_id + ".value"] = auto;
        } else {
          project["items." + f.field_id] = "$items." + f.field_id + "";
        }
      });

      pipe.push({
        $project: project,
      });

      const cursor = ic.find();
      change[ds.datastore_id] = 0;

      while (await cursor.hasNext()) {
        const item = await cursor.next();
        /* 查询条件 */
        const query = { _id: item._id };
        await ic.findOneAndUpdate(query, pipe);
        change[ds.datastore_id]++;
      }
    }
  }

  return change;
}
