/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reaper

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/koderover/zadig/pkg/microservice/reaper/internal/s3"
	"github.com/koderover/zadig/pkg/tool/log"
)

// 上传用户文件到s3
func (r *Reaper) archiveS3Files() (err error) {
	if r.Ctx.FileArchiveCtx != nil && r.Ctx.StorageURI != "" {
		var store *s3.S3

		if store, err = s3.NewS3StorageFromEncryptedURI(r.Ctx.StorageURI); err != nil {
			log.Errorf("failed to create s3 storage %s", r.Ctx.StorageURI)
			return
		}
		if store.Subfolder != "" {
			store.Subfolder = fmt.Sprintf("%s/%s/%d/%s", store.Subfolder, r.Ctx.PipelineName, r.Ctx.TaskID, "file")
		} else {
			store.Subfolder = fmt.Sprintf("%s/%d/%s", r.Ctx.PipelineName, r.Ctx.TaskID, "file")
		}

		src := filepath.Join(r.ActiveWorkspace, r.Ctx.FileArchiveCtx.FileLocation, r.Ctx.FileArchiveCtx.FileName)
		err = s3.Upload(
			context.Background(),
			store,
			src,
			r.Ctx.FileArchiveCtx.FileName,
		)

		if err != nil {
			log.Errorf("failed to upload package %s, %v", src, err)
			return err
		}
	} else {
		return r.archiveTestFiles()
	}

	return nil
}

func (r *Reaper) archiveTestFiles() (err error) {
	if r.Ctx.Archive != nil && r.Ctx.StorageURI != "" {
		var store *s3.S3

		if store, err = s3.NewS3StorageFromEncryptedURI(r.Ctx.StorageURI); err != nil {
			log.Errorf("failed to create s3 storage %s", r.Ctx.StorageURI)
			return
		}
		fileType := "test"
		if strings.Contains(r.Ctx.Archive.File, ".tar.gz") {
			fileType = "file"
		}
		if store.Subfolder != "" {
			store.Subfolder = fmt.Sprintf("%s/%s/%d/%s", store.Subfolder, r.Ctx.PipelineName, r.Ctx.TaskID, fileType)
		} else {
			store.Subfolder = fmt.Sprintf("%s/%d/%s", r.Ctx.PipelineName, r.Ctx.TaskID, fileType)
		}

		filePath := path.Join(r.Ctx.Archive.Dir, r.Ctx.Archive.File)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// no file found, skipped
			//log.Warningf("upload filepath not exist")
			return nil
		}

		err = s3.Upload(
			context.Background(),
			store,
			filePath,
			r.Ctx.Archive.File,
		)

		if err != nil {
			log.Errorf("failed to upload package %s, %v", filePath, err)
			return err
		}
	}

	return nil
}