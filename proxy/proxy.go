package proxy

import (
    "context"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
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
    
    // 创建所有必要的注册器
    endpointRegistry := adapter.NewRegistry[adapter.Endpoint]()
    inboundRegistry := adapter.NewRegistry[adapter.Inbound]()
    outboundRegistry := adapter.NewRegistry[adapter.Outbound]()
    serviceRegistry := service.NewRegistry() 

    // 使用 sing-box 的 Context 函数进行正确的注册
    ctx = box.Context(ctx, inboundRegistry, outboundRegistry, endpointRegistry)
    
    // 注册 service registry
    ctx = service.ContextWith[service.Registry](ctx, serviceRegistry)
    
    // 添加默认服务注册
    ctx = service.ContextWithDefaultRegistry(ctx)
    
    // 创建 box 实例
    instance, err := box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        return E.Cause(err, "create client")
    }
    
    // Start 实例
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