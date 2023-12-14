package user

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/auth/app"
	"github.com/visonlv/iot-engine/auth/model"
	"github.com/xuri/excelize/v2"
)

func TestMetadataSet(t *testing.T) {
	logger.Infof("====%s %s", utilsx.Md5Encode("1234qwer"), EncodePassword(utilsx.Md5Encode("1234qwer")))
}

func TestImportExcel(t *testing.T) {
	app.Init("D:\\GoWork\\src\\git.infore-robotics.cn\\service-robotics-department-2\\auth\\config.toml")

	list, err := model.UserList(nil)
	if err != nil {
		logger.Infof("UserList :%s", err.Error())
		return
	}
	userMap := make(map[string]*model.UserModel, 0)
	for _, v := range list {
		userMap[v.Phone] = v
	}

	fileData, _ := os.ReadFile("C:\\Users\\lvsx1\\Documents\\WXWork\\1688857119421218\\Cache\\File\\2023-03\\花名册20230322-to吕松鑫（用于企微聊天机器人配置）.xlsx")
	reader := bytes.NewBuffer(fileData)
	f, err := excelize.OpenReader(reader)
	if err != nil {
		logger.Infof("open file fail err:%s", err)
		return
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return
	}

	id := 21
	rows, err := f.GetRows("sheet1")
	for i := 0; i < len(rows); i++ {
		if i == 0 {
			continue
		}
		if rows[i][0] == "" {
			continue
		}
		nickName := rows[i][0]
		name := rows[i][3]
		phone := rows[i][4]
		email := fmt.Sprintf("%s@infore.com", name)

		logger.Infof("rows[i] i:%d nickName:%s name:%s phone:%s email:%s", i, nickName, name, phone, email)

		m, ok := userMap[phone]
		if ok {
			// update
			m.Account = name
			m.Phone = phone
			m.NickName = nickName
			m.Email = email
			_, err := model.UserUpdate(nil, m)
			if err != nil {
				logger.Infof("update fail phone:%s err:%s", phone, err.Error())
				return
			}
		} else {
			// add
			m = &model.UserModel{
				Id:       fmt.Sprintf("%d", id),
				Account:  name,
				Password: "0e5e251a746e3c48694707f227282ba8",
				NickName: nickName,
				Email:    email,
				Phone:    phone,
			}
			err := model.UserAdd(nil, m)
			if err != nil {
				logger.Infof("add fail phone:%s err:%s", phone, err.Error())
				return
			}
			id++
		}

	}
}
