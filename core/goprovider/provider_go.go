package goprovider

import "duwhy/core"

// Not Implement
type GoProvider struct{}

func (gp *GoProvider) GetInfoByPath(pathname string, opts *core.InfoOption) (*core.InfoItem, error) {
	return nil, nil
}
