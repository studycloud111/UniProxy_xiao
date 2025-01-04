package proxy

import (
    "context"

    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing/service"
    "github.com/sagernet/sing-box/adapter"
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
    c, err := GetSingBoxConfig(uuid, server)
    if err != nil {
        return err
    }

    // 创建基础 context
    ctx := context.Background()
    // 创建默认 registry
    ctx = service.ContextWithDefaultRegistry(ctx)
    
    // 创建并注册必要的 registry
    inboundRegistry := service.NewRegistry()
    outboundRegistry := service.NewRegistry()
    endpointRegistry := service.NewRegistry()
    
    // 使用 sing-box 的 Context 函数设置 registry
    ctx = box.Context(ctx, inboundRegistry, outboundRegistry, endpointRegistry)

    // 创建 box 实例
    client, err = box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        return err
    }

    err = client.Start()
    if err != nil {
        return err
    }

    Running = true
    return nil
}

func StopProxy() {
    if Running {
        if client != nil {
            client.Close()
            client = nil
        }
        Running = false
    }
}

func ClearSystemProxy() error {
    if Running {
        if client != nil {
            client.Close()
            client = nil
        }
        Running = false
    }
    sysproxy.ClearSystemProxy()
    return nil
}