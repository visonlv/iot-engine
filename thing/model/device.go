package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 设备表
var (
	device_model = &DeviceModel{}
)

type DeviceModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Pk     string `gorm:"type:varchar(64);comment:产品pk"`
	Name   string `gorm:"type:varchar(64);comment:设备名称"`
	Sn     string `gorm:"type:varchar(64);comment:设备Sn"`
	Group  int32  `gorm:"type:int(11);comment:分组标识"`
	Secret string `gorm:"type:varchar(64);comment:秘钥"`
	Desc   string `gorm:"type:varchar(64);comment:描述"`
}

func (a *DeviceModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*DeviceModel) TableName() string {
	return "t_device"
}

func DeviceAdd(tx *mysqlx.MysqlClient, m *DeviceModel) error {
	if err := getTx(tx).Model(device_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func DeviceGet(tx *mysqlx.MysqlClient, id string) (*DeviceModel, error) {
	item := &DeviceModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func DeviceGetBySn(tx *mysqlx.MysqlClient, sn string) (*DeviceModel, error) {
	item := &DeviceModel{}
	result := getTx(tx).Where("sn = ? AND is_delete = ?", sn, 0).First(item)
	return item, result.GetDB().Error
}

func DeviceDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(device_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func DeviceUpdate(tx *mysqlx.MysqlClient, m *DeviceModel) error {
	return getTx(tx).UpdateEx(m)
}

func DevicePage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, pk, name, sn string) ([]*DeviceModel, int32, error) {
	items := []*DeviceModel{}
	var total int32
	query := getTx(tx).Model(device_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if pk != "" {
		query = query.Where("pk = ?", pk)
	}
	if sn != "" {
		query = query.Where("sn = ?", sn)
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func DeviceList(tx *mysqlx.MysqlClient, pk, name, sn string) (list []*DeviceModel, err error) {
	query := getTx(tx).Model(device_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if pk != "" {
		query = query.Where("pk = ?", pk)
	}
	if sn != "" {
		query = query.Where("sn = ?", sn)
	}

	err = query.FindList(&list)
	return
}

func DeviceCountByPk(tx *mysqlx.MysqlClient, pk string) (count int64, err error) {
	result := getTx(tx).Model(device_model).Where("pk = ? AND is_delete = ?", pk, 0).Count(&count)
	err = result.GetDB().Error
	return
}
