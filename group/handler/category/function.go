package category

import (
	"encoding/json"
	"fmt"

	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/group/model"
	pb "github.com/visonlv/iot-engine/group/proto"
)

func CategoryToCategoryPb(m *model.CategoryModel) (*pb.Category, error) {
	itemRet := &pb.Category{}
	err := utilsx.DeepCopy(m, itemRet)
	if err != nil {
		return nil, err
	}
	itemRet.UpdateTime = m.CreatedAt.UnixMilli()
	return itemRet, nil
}

func IsContentValid(content string) ([]*pb.CategoryNodeItem, error) {
	list := make([]*pb.CategoryNodeItem, 0)
	err := json.Unmarshal([]byte(content), &list)
	if err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, fmt.Errorf("至少需要配置一个节点")
	}
	enterIndex := make(map[string]string, 0)
	enterGroup := make(map[int32]int32, 0)
	for _, v := range list {
		if _, ok := enterIndex[v.Index]; ok {
			return nil, fmt.Errorf("节点索引不可重复配置")
		}
		enterIndex[v.Index] = v.Index

		if v.Start > v.End {
			return nil, fmt.Errorf("end 必须大于等于 start")
		}
		if v.Start < define.MinGroup || v.End > define.MaxGroup {
			return nil, fmt.Errorf("分组范围必须在0-99之间")
		}

		var i int32
		for i = v.Start; i <= v.End; i++ {
			if _, ok := enterGroup[i]; ok {
				return nil, fmt.Errorf("分组不可重复 %d", i)
			}
			enterGroup[i] = i
		}
	}
	if len(enterGroup) != (define.MaxGroup - define.MinGroup + 1) {
		return nil, fmt.Errorf("分组必须全覆盖0-100")
	}
	return list, nil
}
