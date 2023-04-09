package memprovider

import (
	"duwhy/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderMemDefault(t *testing.T) {
	ipb, err := NewMemDuFileBuilder("/Users/long/go/src/duwhy/xx.log")
	assert.Nil(t, err)

	ip, err := ipb.Build()
	assert.Nil(t, err)
	assert.NotNil(t, ip)

	di, err := ip.GetInfoByPath(".", nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(di.Childs))
	assert.Equal(t, 8268, di.SizeKB)

	// provider test with deep
	i, err := ip.GetInfoByPath(".", &core.InfoOption{Deep: 5, MaxItems: 10, LongTailPercent: 1})
	assert.Nil(t, err)
	assert.Equal(t, 5, len(i.Childs))
	assert.Equal(t, 8268, i.SizeKB)
	assert.Equal(t, 1680960180, i.ModifiedTimeStamp) // in sec
	assert.Equal(t, i.Name, ".")

	pytest, err := ip.GetInfoByPath("./pytest", nil)
	assert.Nil(t, err)
	assert.NotNil(t, pytest)
	// item test

	ci, ok := i.GetChildItemByPaths([]string{".", "pytest", "Dockerfile.base"}, false)
	assert.True(t, ok)
	assert.Equal(t, 0, len(ci.Childs))
	assert.Equal(t, 4, ci.SizeKB)
	assert.Equal(t, "Dockerfile.base", ci.Name)
	assert.Equal(t, "./pytest/Dockerfile.base", ci.GetFullName())

	nci, ok := i.GetChildItem("./pytest/Dockerfile.base")
	assert.False(t, ok)
	assert.Nil(t, nci)

	ni, ok := i.GetChildItem("pytest")
	assert.True(t, ok)
	assert.Equal(t, 5, len(ni.Childs))
	assert.Equal(t, 8160, ni.SizeKB)
	assert.Equal(t, "pytest", ni.Name)
	assert.Equal(t, "./pytest", ni.GetFullName())
}

func TestProviderMemBase(t *testing.T) {
	ipb, err := NewMemDuFileBuilder("/Users/long/go/src/duwhy/xx.log")
	assert.Nil(t, err)

	ip, err := ipb.Build()
	assert.Nil(t, err)
	assert.NotNil(t, ip)

	di, err := ip.GetInfoByPath(".", nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(di.Childs))
	assert.Equal(t, 8268, di.SizeKB)

	di1, err := ip.GetInfoByPath("./config", &core.InfoOption{
		Deep:            1,
		MaxItems:        5,
		LongTailPercent: 0.95,
	})

	assert.Nil(t, err)
	assert.Equal(t, 6, len(di1.Childs))

	di2, err := ip.GetInfoByPath("./config", &core.InfoOption{
		Deep:            1,
		MaxItems:        5,
		LongTailPercent: 0.5,
	})

	assert.Nil(t, err)
	assert.Equal(t, 4, len(di2.Childs))

}
