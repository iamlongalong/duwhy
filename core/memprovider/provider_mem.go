package memprovider

import (
	"bufio"
	"bytes"
	"duwhy/core"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/pkg/errors"
)

type memProviderSourceType int

var (
	MemProviderTypeDUFile memProviderSourceType = 1

	// not implement yet
	MemProviderTypeDuComman memProviderSourceType = 2
)

type MemDUBuilderOption struct {
	Ignore []string
}

func NewMemDuFileBuilder(dufile string, opts *MemDUBuilderOption) (core.IProviderBuilder, error) {
	if opts == nil {
		opts = &MemDUBuilderOption{Ignore: make([]string, 0)}
	}

	ignores := make([]string, 0, len(opts.Ignore))

	for _, ig := range opts.Ignore {
		ig = strings.TrimSuffix(ig, "*")

		if !strings.HasPrefix(ig, "./") {
			if !strings.HasPrefix(ig, "/") {
				ig = "./" + ig
			}
		}

		if !strings.HasSuffix(ig, "/") {
			ig = ig + "/"
		}

		ignores = append(ignores, ig)
	}

	f, err := os.OpenFile(dufile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "create mem provider builder fail open dufile")
	}

	return &MemBuilder{
		SourceType: MemProviderTypeDUFile,
		dufile:     f,
		ignores:    ignores,
	}, nil
}

type MemBuilder struct {
	// dufile or use go
	SourceType memProviderSourceType

	dufile *os.File

	ignores []string
}

func (b *MemBuilder) Build() (core.IProvider, error) {
	var ip core.IProvider

	switch b.SourceType {
	case MemProviderTypeDUFile:
		var err error
		ip, err = buildFromDuFile(b.dufile, true, b.ignores)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid mem provider source type")
	}

	return ip, nil
}

func buildFromDuFile(f *os.File, omitParseErr bool, ignores []string) (*MemProvider, error) {

	r := bufio.NewReader(f)
	defer f.Close()

	rootItem := core.NewInfoItem()

	var lastParentItem *core.InfoItem
	var lastParentPath string

readnewline:
	for {
		// 姑且认为一行的长度不会超出
		l, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if !omitParseErr {
				return nil, errors.Wrap(err, "fail to read line")
			}

			log.Println(errors.Wrap(err, "fail to read line"))
			continue
		}

		nl := make([]byte, len(l))
		copy(nl, l)

		duline, err := parseDuLine(nl)
		if err != nil {
			if !omitParseErr {
				return nil, errors.Wrap(err, "parse du line fail")
			}

			log.Println(errors.Wrap(err, "fail to read line"))
			continue
		}

		// check if should ignore
		for _, ig := range ignores {
			if strings.HasPrefix(strings.Join(duline.Paths, "/"), ig) {
				continue readnewline
			}
		}

		tarNode := core.NewInfoItem()
		bindDuLineToInfoItem(duline, tarNode)

		if len(duline.Paths) == 0 {
			if !omitParseErr {
				return nil, errors.Errorf("invalid duline paths of zero")
			}

			log.Println("invalid duline paths of zero")
			continue
		}

		if len(duline.Paths) == 1 {
			// root
			rootItem.Name = duline.Paths[0]
			rootItem.ModifiedTimeStamp = int(duline.ModifiedTime.Unix())
			rootItem.SizeKB = duline.SizeKB
		} else {

			parentPath := duline.Paths[0 : len(duline.Paths)-1]

			parentPathStr := strings.Join(parentPath, "/")
			if lastParentPath != parentPathStr {
				lastParentItem, _ = rootItem.GetChildItemByPaths(parentPath, true)
				lastParentPath = parentPathStr
			}

			lastParentItem.AddChildItem(tarNode, false, true)
		}

	}

	cache := gcache.New(1000).LRU().Expiration(time.Minute * 5).Build()

	memprovider := &MemProvider{
		root:       rootItem,
		pathsCache: cache,
	}

	return memprovider, nil
}

func bindDuLineToInfoItem(l DULine, ii *core.InfoItem) {
	ii.Name = l.Paths[len(l.Paths)-1]
	ii.ModifiedTimeStamp = int(l.ModifiedTime.Unix())
	ii.SizeKB = l.SizeKB
}

type DULine struct {
	SizeKB       int
	ModifiedTime time.Time
	Paths        []string
}

func parseDuLine(b []byte) (DULine, error) {
	l := DULine{}

	infos := bytes.Split(b, []byte("\t"))
	if len(infos) != 3 {
		return l, errors.Errorf("line block num not 3 : %s", string(b))
	}

	var err error
	l.SizeKB, err = strconv.Atoi(string(infos[0]))
	if err != nil {
		return l, errors.Wrapf(err, "parse size fail : %s", string(b))
	}

	l.ModifiedTime, err = time.Parse("2006-01-02 15:04", string(infos[1]))
	if err != nil {
		return l, errors.Wrapf(err, "parse modified time fail : %s", string(b))
	}

	ps := bytes.Split(infos[2], []byte("/"))
	l.Paths = make([]string, 0, len(ps))

	for _, p := range ps {
		l.Paths = append(l.Paths, string(p))
	}

	return l, nil
}

type MemProvider struct {
	pathsCache gcache.Cache

	root *core.InfoItem
}

var dfOption = &core.InfoOption{
	Deep:            0, // default with no children
	LongTailPercent: 1, // default one of itself
	MaxItems:        1, // default one of itself
}

func (mp *MemProvider) GetInfoByPath(pathname string, opts *core.InfoOption) (*core.InfoItem, error) {
	{
		if opts == nil {
			opts = dfOption
		}
	}

	var err error

	var startItem *core.InfoItem

	if opts == dfOption {
		isi, err := mp.pathsCache.Get(pathname)
		if err == nil {
			startItem = isi.(*core.InfoItem)
		}
	}

	if startItem == nil {
		ps := strings.Split(filepath.Clean(pathname), string(filepath.Separator))

		startItem = mp.root
		var ok bool
		for _, p := range ps {
			startItem, ok = startItem.GetChildItem(p)
			if !ok {
				return nil, core.ErrPathNotExist
			}
		}

		if opts == dfOption {
			mp.pathsCache.Set(pathname, startItem) // 姑且这么设置吧，之后看看使用情况
		}
	}

	ii := startItem.Clone(opts.Deep, nil, false)

	err = core.FilterChildrens(ii, opts.MaxItems, opts.LongTailPercent)
	if err != nil {
		return nil, err
	}
	ii.PercentOfParent = 10000

	return ii, err
}
