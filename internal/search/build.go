package search

import (
	"context"
	"path"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	log "github.com/sirupsen/logrus"
)

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int) error {
	var objCount uint64 = 0
	var (
		err error
		fi  model.Obj
	)
	defer func() {
		now := time.Now()
		eMsg := ""
		if err != nil {
			eMsg = err.Error()
		} else {
			log.Infof("success build index, count: %d", objCount)
		}
		WriteProgress(&model.IndexProgress{
			FileCount:    objCount,
			IsDone:       err == nil,
			LastDoneTime: &now,
			Error:        eMsg,
		})
	}()
	admin, err := db.GetAdmin()
	if err != nil {
		return err
	}
	for _, indexPath := range indexPaths {
		walkFn := func(indexPath string, info model.Obj, err error) error {
			for _, avoidPath := range ignorePaths {
				if indexPath == avoidPath {
					return filepath.SkipDir
				}
			}
			// ignore root
			if indexPath == "/" {
				return nil
			}
			err = instance.Index(ctx, path.Dir(indexPath), info)
			if err != nil {
				return err
			} else {
				objCount++
			}
			if objCount%100 == 0 {
				log.Infof("index obj count: %d", objCount)
				log.Debugf("current success index: %s", indexPath)
				WriteProgress(&model.IndexProgress{
					FileCount:    objCount,
					IsDone:       false,
					LastDoneTime: nil,
				})
			}
			return nil
		}
		fi, err = fs.Get(ctx, indexPath)
		if err != nil {
			return err
		}
		// TODO: run walkFS concurrently
		err = fs.WalkFS(context.WithValue(ctx, "user", admin), maxDepth, indexPath, fi, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}

func Clear(ctx context.Context) error {
	return instance.Clear(ctx)
}
