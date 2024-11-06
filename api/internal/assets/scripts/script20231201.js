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
            let n = 3;
            if (datastore.api_key === "paymentInterest" || datastore.api_key === "repayment") {
                acount++;
                /* 查找顾客下的的所有字段 */
                const fsc = db.collection("fields")
                const fs = await fsc.find().toArray();

                /* 循环所有字段，更新数据 */
                for (let m = 0; m < fs.length; m++) {
                    const field = fs[m];
                    if (field.datastore_id === datastore.datastore_id && field.field_id !== "year" && field.field_id !== "month") {
                        await fsc.updateOne(
                            { datastore_id: datastore.datastore_id ,field_id: field.field_id },
                            {$set:{
                                'display_order' : n,
                            }    
                            }
                        );
                        n++;
                    }
                }

                await fsc.updateOne(
                    { datastore_id: datastore.datastore_id ,field_id: "year"  },
                    {$set:{
                        'display_order' :1,
                    }    
                    }
                );
                await fsc.updateOne(
                    { datastore_id: datastore.datastore_id ,field_id: "month"  },
                    {$set:{
                        'display_order' :2,
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