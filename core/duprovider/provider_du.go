package duprovider

import "duwhy/core"

// Not Implement
type DuProvider struct{}

func (gp *DuProvider) GetInfoByPath(pathname string, opts *core.InfoOption) (*core.InfoItem, error) {
	return nil, nil
}
