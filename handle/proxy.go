package handle

import (
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "github.com/studycloud111/UniProxy_xiao/proxy"
)

type StartUniProxyRequest struct {
    Tag        string `json:"tag"`
    Uuid       string `json:"uuid"`
    GlobalMode bool   `json:"global_mode"`
}

func StartUniProxy(c *gin.Context) {
    p := StartUniProxyRequest{}
    err := c.ShouldBindJSON(&p)
    if err != nil {
        log.WithError(err).Error("Failed to bind request")
        c.JSON(400, Rsp{
            Success: false,
            Message: err.Error(),
        })
        return
    }
    
    // 检查 server 是否存在
    server, exists := servers[p.Tag]
    if !exists {
        log.WithField("tag", p.Tag).Error("Server not found")
        c.JSON(400, Rsp{
            Success: false,
            Message: "server not found",
        })
        return
    }

    proxy.GlobalMode = p.GlobalMode
    err = proxy.StartProxy(p.Tag, p.Uuid, server)  // 使用找到的 server
    if err != nil {
        log.WithFields(log.Fields{
            "error": err,
            "tag": p.Tag,
            "global_mode": p.GlobalMode,
        }).Error("Failed to start proxy")
        
        c.JSON(500, Rsp{
            Success: false,
            Message: err.Error(),
        })
        return
    }
    
    c.JSON(200, Rsp{
        Success: true,
        Message: "ok",
        Data: StatusData{
            Inited:      inited,
            Running:     proxy.Running,
            GlobalMode:  proxy.GlobalMode,
            SystemProxy: proxy.SystemProxy,
        },
    })
}

func StopUniProxy(c *gin.Context) {
    if proxy.Running {
        proxy.StopProxy()
    }
    c.JSON(200, Rsp{
        Success: true,
        Message: "ok",
    })
}

func SetSystemProxy(c *gin.Context) {
    c.JSON(200, Rsp{
        Success: true,
        Message: "ok",
    })
}

func ClearSystemProxy(c *gin.Context) {
    err := proxy.ClearSystemProxy()
    if err != nil {
        c.JSON(500, Rsp{
            Success: false,
            Message: err.Error(),
        })
        return
    }
    c.JSON(200, Rsp{
        Success: true,
        Message: "ok",
    })
}