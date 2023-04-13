package core

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

var (
	ErrPathNotExist = os.ErrNotExist
)

func NewInfoItem() *InfoItem {
	return &InfoItem{Childs: make([]*InfoItem, 0), childmaps: map[string]*InfoItem{}}
}

type InfoItem struct {
	Name              string
	SizeKB            int
	ModifiedTimeStamp int

	PercentOfParent int // data * 10000, for 82.1% is 8210

	Childs []*InfoItem

	// calcedSizeKB int
	parent    *InfoItem
	childmaps map[string]*InfoItem

	sorted bool
}

// AddChildItemTo 添加到指定路径下
// params: forceCover , set child anyway even if child exists
// params: updateInfo, if child exists, then update SizeKB and ModifiedTimeStamp
func (ii *InfoItem) AddChildItemTo(paths []string, ci *InfoItem, forceCover bool, updateInfo bool) (alreadyHasItem bool) {
	targetParentItem, _ := ii.GetChildItemByPaths(paths, true)

	return targetParentItem.AddChildItem(ci, forceCover, updateInfo)
}

func (ii *InfoItem) AddChildItem(ci *InfoItem, forceCover bool, updateInfo bool) (alreadyHasItem bool) {
	alreadyTarget, ok := ii.GetChildItem(ci.Name)
	if !ok {
		ii.Childs = append(ii.Childs, ci)
		ii.childmaps[ci.Name] = ci
		ci.parent = ii

		ii.sorted = false
		return false
	}

	if forceCover {
		for i, c := range ii.Childs {
			if c.Name == ci.Name {
				ii.Childs[i] = ci
			}
		}

		ii.childmaps[ci.Name] = ci
		ci.parent = ii
		ii.sorted = false
		return true
	}

	if updateInfo {
		alreadyTarget.ModifiedTimeStamp = ci.ModifiedTimeStamp
		alreadyTarget.SizeKB = ci.SizeKB
		ii.sorted = false
		return true
	}

	return ok
}

// MustGetSize 若未设置当前路径的大小，则获取子路径的大小之和
func (ii *InfoItem) MustGetSize(minFileSizeKB int) int {
	if ii.SizeKB != 0 {
		return ii.SizeKB
	}

	for _, c := range ii.Childs {
		ii.SizeKB += c.MustGetSize(minFileSizeKB)
	}

	if ii.SizeKB == 0 {
		ii.SizeKB = minFileSizeKB
	}

	return ii.SizeKB
}

func (ii *InfoItem) SortChildren() {
	if ii.sorted {
		return
	}

	if len(ii.Childs) == 0 {
		return
	}

	// on reverse order
	sort.Slice(ii.Childs, func(i, j int) bool {
		return ii.Childs[i].SizeKB > ii.Childs[j].SizeKB
	})

	ii.sorted = true
}

func (ii *InfoItem) GetChildItemByPaths(paths []string, create bool) (item *InfoItem, ok bool) {
	lastItem := ii

	for _, p := range paths {
		titem, ok := lastItem.GetChildItem(p)
		if !ok { // 不存在
			if !create {
				return nil, false
			}

			titem = NewInfoItem()
			titem.Name = p
			lastItem.AddChildItem(titem, false, true)

			lastItem = titem
			continue
		}

		lastItem = titem
	}

	return lastItem, true
}

// GetChildItem by name
// !Attention name="." and name="" both means current item
func (ii *InfoItem) GetChildItem(name string) (*InfoItem, bool) {
	if name == "." || name == "" {
		return ii, true
	}

	if ii.childmaps != nil {
		cii, ok := ii.childmaps[name]
		return cii, ok
	}

	for _, ci := range ii.Childs {
		if ci.Name == name {
			return ci, true
		}
	}

	return nil, false
}

func (ii *InfoItem) Clone(deepth int, parent *InfoItem, withInternalFields bool) *InfoItem {
	nii := *ii

	if deepth == 0 {
		nii.Childs = make([]*InfoItem, 0)
		nii.childmaps = map[string]*InfoItem{}
		return &nii
	}

	nii.Childs = make([]*InfoItem, len(nii.Childs))
	nii.parent = parent

	if withInternalFields {
		nii.childmaps = make(map[string]*InfoItem, len(nii.Childs))
	} else {
		nii.childmaps = nil
	}

	for i, child := range ii.Childs {
		nc := child.Clone(deepth-1, &nii, withInternalFields)
		nii.Childs[i] = nc

		if withInternalFields {
			nii.childmaps[nc.Name] = nc
		}
	}

	return &nii
}

func (ii *InfoItem) GetFullName() string {
	namestack := make([]string, 0, 1)

	ni := ii
	for ni != nil {
		namestack = append(namestack, ni.Name)

		ni = ni.parent
	}

	tnamestack := make([]byte, 0, len(namestack)*6)
	for i := len(namestack); i > 0; i-- {
		tnamestack = append(tnamestack, []byte(namestack[i-1])...)
		if i-1 != 0 {
			tnamestack = append(tnamestack, filepath.Separator)
		}
	}

	return string(tnamestack)
}

func FilterChildrens(ii *InfoItem, maxItem int, longTailPercent float64) error {
	if longTailPercent > 1 || longTailPercent <= 0 {
		return errors.New("long tail percent must be in 0 < x <= 1")
	}
	ii.MustGetSize(4) // 默认最小 file block 为 4kb

	ii.SortChildren()

	nowsize := 0
	nowItems := 0

	childs := make([]*InfoItem, 0, maxItem)

	for _, ci := range ii.Childs {

		// 数量超出
		if maxItem != 0 && nowItems == maxItem {
			otherItem := &InfoItem{
				Name:            "Others",
				SizeKB:          ii.SizeKB - nowsize,
				PercentOfParent: (ii.SizeKB - nowsize) * 10000 / ii.SizeKB,
			}
			childs = append(childs, otherItem)
			break
		}

		// 大小超出
		if float64(nowsize+ci.SizeKB)/float64(ii.SizeKB) > longTailPercent {
			if len(childs) != 0 { // 一个子节点都没有就跳过
				otherItem := &InfoItem{
					Name:            "Others",
					SizeKB:          ii.SizeKB - nowsize,
					PercentOfParent: (ii.SizeKB - nowsize) * 10000 / ii.SizeKB,
				}
				childs = append(childs, otherItem)
				break
			}
		}

		ci.PercentOfParent = (ci.SizeKB * 10000) / ii.SizeKB

		childs = append(childs, ci)
		nowsize += ci.SizeKB
		nowItems++
	}

	ii.Childs = childs

	for _, c := range childs {
		err := FilterChildrens(c, maxItem, longTailPercent)
		if err != nil {
			return err
		}
	}

	return nil
}

type InfoOption struct {
	MaxItems        int     // 最大 item 数量
	LongTailPercent float64 // 长尾的切割百分比，例如 95% 意味着排序的后 5% 会被放入 OtherInfo

	Deep int // 深入子节点
}

type IProviderBuilder interface {
	Build() (IProvider, error)
}

type IProvider interface {
	GetInfoByPath(pathname string, opts *InfoOption) (*InfoItem, error)
}
