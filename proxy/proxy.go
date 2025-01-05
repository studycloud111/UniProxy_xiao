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
    ctx := service.ContextWithDefaultRegistry(context.Background())
    
    // 创建默认注册器
    inboundReg := new(experimental.InboundRegistry)
    outboundReg := new(experimental.OutboundRegistry)
    endpointReg := new(experimental.EndpointRegistry)
    
    ctx = box.Context(ctx, inboundReg, outboundReg, endpointReg)
    
    // 创建 box 实例
    instance, err := box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        return E.Cause(err, "create client")
    }
    
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