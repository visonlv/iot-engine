package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"gorm.io/gorm"
)

// 影子表
var (
	shadow_model = &ShadowModel{}
)

type ShadowModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Sn          string `gorm:"type:varchar(64);comment:设备sn"`
	PSn         string `gorm:"type:varchar(64);comment:父设备sn"`
	Group       int32  `gorm:"type:int(11);comment:分组标识"`
	Pk          string `gorm:"type:varchar(64);comment:产品KEY"`
	Shadow      string `gorm:"type:text;comment:影子信息"`
	LastVersion int64  `gorm:"type:bigint(20);comment:最后版本号"`
}

func (a *ShadowModel) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (*ShadowModel) TableName() string {
	return "t_shadow"
}

func ShadowAdd(tx *mysqlx.MysqlClient, m *ShadowModel) error {
	if err := getTx(tx).Model(shadow_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func ShadowGet(tx *mysqlx.MysqlClient, id string) (*ShadowModel, error) {
	item := &ShadowModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func ShadowGetBySn(tx *mysqlx.MysqlClient, deviceSn string) (*ShadowModel, error) {
	item := &ShadowModel{}
	result := getTx(tx).Where("sn = ? AND is_delete = ?", deviceSn, 0).First(item)
	return item, result.GetDB().Error
}

func ShadowDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(shadow_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func ShadowUpdate(tx *mysqlx.MysqlClient, m *ShadowModel) error {
	return getTx(tx).UpdateEx(m)
}

func ShadowPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32) ([]*ShadowModel, int32, error) {
	items := []*ShadowModel{}
	var total int32
	err := getTx(tx).Model(shadow_model).Where("is_delete = ?", 0).Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func ShadowList(tx *mysqlx.MysqlClient) (list []*ShadowModel, err error) {
	err = getTx(tx).Model(shadow_model).Where("is_delete = ?", 0).FindList(&list)
	return
}

func ShadowListBetweenGroupAndPk(tx *mysqlx.MysqlClient, pk []string, start, end int32) (list []*ShadowModel, err error) {
	query := getTx(tx).Model(shadow_model).Where("'group' >= ? and 'group' <= ? and is_delete = ?", start, end, 0)
	if pk != nil {
		query.Where("pk in ?", pk)
	}
	err = query.FindList(&list)
	return
}

func ShadowListInSn(tx *mysqlx.MysqlClient, sns []string) (list []*ShadowModel, err error) {
	err = getTx(tx).Model(shadow_model).Where("sn in ? and is_delete = ?", sns, 0).FindList(&list)
	return
}

func ShadowUpdateBySnAndShadowAndVersion(tx *mysqlx.MysqlClient, sn, shadow string, version int64) error {
	result := getTx(tx).Model(shadow_model).Where("is_delete = ? AND sn = ?", 0, sn).Update("shadow", shadow).Update("last_version", version)
	return result.GetDB().Error
}

func ShadowUpdateByIdAndPsn(tx *mysqlx.MysqlClient, id, pSn string) error {
	result := getTx(tx).Model(shadow_model).Where("is_delete = ? AND id = ?", 0, id).Update("p_sn", pSn)
	return result.GetDB().Error
}
