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
            if (datastore.api_key === "paymentInterest") {
                /* 查找顾客下的的权限 */
                const pe = db.collection("permissions");
                const ps = await pe.find().toArray();
                for (let m = 0; m < ps.length; m++) {
                    const permission = ps[m];
                    if (permission.action_type === "datastore" && permission.app_id === datastore.app_id) {
                        acount++;
                        const pas = permission.actions;
                        for (let n = 0; n < pas.length; n++) {
                            const pa = pas[n]
                            if (pa && pa.object_id === datastore.datastore_id) {
                                let arr = pa.fields
                                let index = arr.indexOf("paymentymd")
                                if(index!=-1){
                                    arr.splice(index,1);
                                }
                                pa.fields = arr
                            } 
                        }
                        await pe.updateOne(
                            { action_type:"datastore" , app_id:datastore.app_id , role_id:permission.role_id},
                            {$set:{
                                "actions" :pas,
                            }
                            }
                        );
                    }
                }
            }
            if (datastore.api_key === "repayment") {
                /* 查找顾客下的的权限 */
                const pe = db.collection("permissions");
                const ps = await pe.find().toArray();
                for (let m = 0; m < ps.length; m++) {
                    const permission = ps[m];
                    if (permission.action_type === "datastore" && permission.app_id === datastore.app_id) {
                        acount++;
                        const pas = permission.actions;
                        for (let n = 0; n < pas.length; n++) {
                            const pa = pas[n]
                            if (pa && pa.object_id === datastore.datastore_id) {
                                let arr = pa.fields
                                let index = arr.indexOf("syokyakuymd")
                                if(index!=-1){
                                    arr.splice(index,1);
                                }
                                pa.fields = arr
                            } 
                        }
                        await pe.updateOne(
                            { action_type:"datastore" , app_id:datastore.app_id , role_id:permission.role_id},
                            {$set:{
                                "actions" :pas,
                            }
                            }
                        );
                    }
                }
            }
        }
    }
    console.log("结束")
    change["処理概要-データまとめ"] = "データ更新の総数：" + acount;
    change["処理概要-終"] = "処理が正常に終了しました";
    
    return change
}