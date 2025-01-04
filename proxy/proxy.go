package proxy

import (
    "context"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/log"
    "github.com/studycloud111/UniProxy_xiao/common/sysproxy"
    "github.com/studycloud111/UniProxy_xiao/v2b"
)

var (
    Running     bool
    SystemProxy bool
    GlobalMode  bool
    TunMode     bool
    InPort      int
    DataPath    string
    ResUrl      string
)

var client *box.Box

func StartProxy(tag string, uuid string, server *v2b.ServerInfo) error {
    if Running {
        StopProxy()
    }
    SystemProxy = true
    
    // 获取配置
    c, err := GetSingBoxConfig(uuid, server)
    if err != nil {
        return err
    }
    
    // 创建带有 logger 的 context
    logger := log.NewDefaultLogger()
    ctx := context.Background()
    ctx = log.ContextWithLogger(ctx, logger)
    
    // 创建客户端
    client, err = box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        return err
    }
    
    // 启动服务
    err = client.Start()
    if err != nil {
        return err
    }
    
    Running = true
    return nil
}

func StopProxy() {
    if Running {
        client.Close()
        Running = false
    }
}

func ClearSystemProxy() error {
    if Running {
        client.Close()
        Running = false
        return nil
    }
    sysproxy.ClearSystemProxy()
    return nil
}