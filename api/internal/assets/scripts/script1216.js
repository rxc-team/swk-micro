async function run() {
  await client.connect();

  const pit = client.db("pit");
  const cc = pit.collection("customers");

  const customers = await cc.find().toArray();

  const change = {};
  /* 顾客为空的场合 */
  if (!customers || customers.length === 0) {
    return change;
  }

  /* 循环所有顾客，更新数据 */
  for (let i = 0; i < customers.length; i++) {
    const cs = customers[i];

    const db = client.db(`pit_${cs.customer_id}`);
    /* 查找顾客下的所有语言 */
    const lc = db.collection("languages");
    const languages = await lc.find().toArray();
    /* 查找顾客下的所有流程 */
    const wfc = db.collection("wf_workflows");
    const workflows = await wfc.find().toArray();

    /* 流程为空的场合，不需要改变 */
    if (!workflows || workflows.length === 0) {
      continue;
    }

    for (let j = 0; j < languages.length; j++) {
      const lan = languages[j];

      const setData = {};
      const unsetData = {};

      for (let k = 0; k < workflows.length; k++) {
        const wf = workflows[k];

        const wsetData = {};

        if (lan.common?.workflows?.hasOwnProperty(wf.wf_id)) {
          const name = lan.common.workflows[wf.wf_id] || "";
          setData[`apps.${wf.app_id}.workflows.${wf.wf_id}`] = name;
          unsetData[`common.workflows.${wf.wf_id}`] = "";
        }

        if (lan.common?.workflows?.hasOwnProperty(`menu_${wf.wf_id}`)) {
          const menu = lan.common.workflows[`menu_${wf.wf_id}`] || "";
          setData[`apps.${wf.app_id}.workflows.menu_${wf.wf_id}`] = menu;
          unsetData[`common.workflows.menu_${wf.wf_id}`] = "";
        }
        /* 第一次的场合，更新流程的key */
        if (j === 0) {
          wsetData["wf_name"] = `apps.${wf.app_id}.workflows.${wf.wf_id}`;
          wsetData[
            "menu_name"
          ] = `apps.${wf.app_id}.workflows.menu_${wf.wf_id}`;

          await wfc.updateOne({ _id: wf._id }, { $set: wsetData });
        }
      }

      await lc.updateOne({ _id: lan._id }, { $set: setData });
      await lc.updateOne({ _id: lan._id }, { $unset: unsetData });
    }
  }

  return change;
}
