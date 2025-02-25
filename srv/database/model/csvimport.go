package model

import (
	"rxcsoft.cn/pit3/srv/database/utils"
)

func getFieldMap(db, appID string) (map[string][]Field, error) {
	param := &FindAppFieldsParam{
		AppID:         appID,
		InvalidatedIn: "true",
	}
	fields, err := FindAppFields(db, param)
	if err != nil {
		utils.ErrorLog("getFieldMap", err.Error())
		return nil, err
	}
	var ds string
	var fs []Field
	result := make(map[string][]Field)
	for index, f := range fields {
		if index == 0 {
			ds = f.DatastoreID
			fs = append(fs, f)

			if len(fields) == 1 {
				result[ds] = fs
			}
			continue
		}

		if len(fields)-1 == index {
			if ds == f.DatastoreID {
				fs = append(fs, f)
				result[ds] = fs
			} else {
				result[ds] = fs
				fs = nil
				ds = f.DatastoreID
				fs = append(fs, f)
				result[ds] = fs
			}
			continue
		}

		if ds == f.DatastoreID {
			fs = append(fs, f)
			continue
		}

		result[ds] = fs
		fs = nil
		ds = f.DatastoreID
		fs = append(fs, f)
	}

	return result, nil
}
