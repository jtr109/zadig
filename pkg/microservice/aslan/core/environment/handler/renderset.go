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

package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	yamlutil "github.com/koderover/zadig/pkg/util/yaml"

	fsservice "github.com/koderover/zadig/pkg/microservice/aslan/core/common/service/fs"

	"github.com/gin-gonic/gin"

	commonservice "github.com/koderover/zadig/pkg/microservice/aslan/core/common/service"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/environment/service"
	internalhandler "github.com/koderover/zadig/pkg/shared/handler"
	e "github.com/koderover/zadig/pkg/tool/errors"
	"github.com/koderover/zadig/pkg/tool/log"
)

func GetServiceRenderCharts(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if c.Query("productName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("productName can not be null!")
		return
	}

	if c.Query("envName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName can not be null!")
		return
	}

	ctx.Resp, ctx.Err = service.GetRenderCharts(c.Query("productName"), c.Query("envName"), c.Query("serviceName"), ctx.Logger)
}

func GetProductDefaultValues(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if c.Query("productName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("productName can not be null!")
		return
	}

	if c.Query("envName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName can not be null!")
		return
	}

	ctx.Resp, ctx.Err = service.GetDefaultValues(c.Query("productName"), c.Query("envName"), ctx.Logger)
}

func CreateOrUpdateRenderChart(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if c.Query("productName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("productName can not be null!")
		return
	}

	if c.Query("envName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName can not be null!")
		return
	}

	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("CreateOrUpdateRenderChart c.GetRawData() err : %v", err)
	}

	args := new(commonservice.RenderChartArg)
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("CreateOrUpdateRenderChart json.Unmarshal err : %v", err)
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
	}
	internalhandler.InsertOperationLog(c, ctx.Username, c.Param("productName"), "新增", "环境变量", c.Query("envName"), string(data), ctx.Logger)

	ctx.Err = service.CreateOrUpdateChartValues(c.Query("productName"), c.Query("envName"), args, ctx.Username, ctx.RequestID, ctx.Logger)
}

func CreateOrUpdateRenderset(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if c.Query("productName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("productName can not be null!")
		return
	}

	if c.Query("envName") == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName can not be null!")
		return
	}

	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("CreateOrUpdateRenderChart c.GetRawData() err : %v", err)
	}

	args := new(commonservice.RendersetArg)
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("CreateOrUpdateRenderChart json.Unmarshal err : %v", err)
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}
	internalhandler.InsertOperationLog(c, ctx.Username, c.Param("productName"), "新增", "环境变量", c.Query("envName"), string(data), ctx.Logger)

	ctx.Err = service.CreateOrUpdateRenderset(c.Query("productName"), c.Query("envName"), args, ctx.Username, ctx.RequestID, ctx.Logger)
}

func GetYamlContent(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	var err error
	codehostID := 0

	codehostIDStr := c.Query("codehostID")
	if len(codehostIDStr) > 0 {
		codehostID, err = strconv.Atoi(codehostIDStr)
		if err != nil {
			ctx.Err = e.ErrInvalidParam.AddDesc("cannot convert codehost id to int")
			return
		}
	}

	if codehostID == 0 && len(c.Query("repoLink")) == 0 {
		ctx.Err = e.ErrInvalidParam.AddDesc("neither codehost nor repo link is specified")
		return
	}

	if len(c.Query("valuesPaths")) == 0 {
		ctx.Err = e.ErrInvalidParam.AddDesc("paths can't be empty")
		return
	}

	pathArr := strings.Split(c.Query("valuesPaths"), ",")

	contentArr := make([][]byte, 0)
	for _, path := range pathArr {
		fileContent, err := fsservice.DownloadFileFromSource(
			&fsservice.DownloadFromSourceArgs{
				CodehostID: codehostID,
				Owner:      c.Query("owner"),
				Repo:       c.Query("repo"),
				Path:       path,
				Branch:     c.Query("branch"),
				RepoLink:   c.Query("repoLink"),
			})
		if err != nil {
			ctx.Err = fmt.Errorf("failed to download file from path %s, err %s", path, err)
			return
		}
		if len(fileContent) > 0 {
			contentArr = append(contentArr, fileContent)
		}
	}

	ret, err := yamlutil.Merge(contentArr)
	if err != nil {
		ctx.Err = fmt.Errorf("failed to merge files, err %s", err)
		return
	}
	ctx.Resp = string(ret)
}
