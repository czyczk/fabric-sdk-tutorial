package controller

import (
	"net/http"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"github.com/gin-gonic/gin"
)

// A ScrewController contains a group name and a `ScrewService`. It also implements the interface `Controller`.
type ScrewController struct {
	GroupName string
	ScrewSvc  *service.ScrewService
}

// GetGroupName returns the group name
func (sc *ScrewController) GetGroupName() string {
	return sc.GroupName
}

// GetEndpointMap implements the interface `Controller` and returns the API endpoints and handlers defined and managed by ScrewController.
func (sc *ScrewController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"/transfer", "POST"}: []gin.HandlerFunc{
			func(c *gin.Context) {
				corpAName := c.PostForm("corpAName")
				corpBName := c.PostForm("corpBName")
				amnt := c.PostForm("amnt")

				// Validity check
				pel := &ParameterErrorList{}

				if corpAName = strings.TrimSpace(corpAName); corpAName == "" {
					*pel = append(*pel, "公司 A 名称不能为空。")
				}

				corpBName = pel.AppendIfEmptyOrBlankSpaces(corpBName, "公司 B 名称不能为空。")

				amntUint := pel.AppendIfNotUint(amnt, "数量必须为正整数。")

				if len(*pel) > 0 {
					c.JSON(http.StatusBadRequest, pel)
					return
				}

				txID, err := sc.ScrewSvc.TransferAndShowEvent(corpAName, corpBName, amntUint)

				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					c.JSON(http.StatusOK, txID)
				}
			},
		},

		urlMethodPair{"/count", "GET"}: []gin.HandlerFunc{
			func(c *gin.Context) {
				corpName := c.Query("corpName")

				// Validity check
				pel := &ParameterErrorList{}

				corpName = pel.AppendIfEmptyOrBlankSpaces(corpName, "公司名称不能为空。")

				if len(*pel) > 0 {
					c.JSON(http.StatusBadRequest, pel)
					return
				}

				count, err := sc.ScrewSvc.Query(corpName)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					// Empty return value indicates invalid query key
					if count == "" {
						*pel = append(*pel, "该公司名称不存在。")
						c.JSON(http.StatusBadRequest, pel)
					} else {
						c.JSON(http.StatusOK, count)
					}
				}
			},
		},
	}
}
