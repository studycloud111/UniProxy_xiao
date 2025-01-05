package proxy

import (
    "context"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/experimental"
    "github.com/sagernet/sing-box/option"
    slog "github.com/sagernet/sing-box/log"
    E "github.com/sagernet/sing/common/exceptions"
    "github.com/sagernet/sing/service"
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
    
    // 创建必要的注册器
    endpointRegistry := new(experimental.EndpointRegistry)
    inboundRegistry := new(experimental.InboundRegistry)
    outboundRegistry := new(experimental.OutboundRegistry)
    
    // 注册到 context
    ctx = service.ContextWith[adapter.EndpointRegistry](ctx, endpointRegistry)
    ctx = service.ContextWith[adapter.InboundRegistry](ctx, inboundRegistry)
    ctx = service.ContextWith[adapter.OutboundRegistry](ctx, outboundRegistry)
    
    // 使用框架提供的 Context 函数
    ctx = box.Context(ctx, inboundRegistry, outboundRegistry, endpointRegistry)
    
    // 创建 box 实例
    instance, err := box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        return E.Cause(err, "create client")
    }
    
    // 启动服务
    err = instance.Start()
    if err != nil {
        instance.Close()
        return E.Cause(err, "start client")
    }
    
    client = instance
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