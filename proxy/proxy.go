package proxy

import (
    "context"
    "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/adapter/outbound"
    "github.com/sagernet/sing-box/adapter/inbound"
    E "github.com/sagernet/sing/common/exceptions"
    "github.com/sagernet/sing/service"
    "github.com/studycloud111/UniProxy_xiao/v2b"
    "github.com/studycloud111/UniProxy_xiao/common/sysproxy"
    log "github.com/sirupsen/logrus"
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

    log.WithFields(log.Fields{
        "tag": tag,
        "server_type": server.Type,
        "server_host": server.Host,
    }).Info("Starting proxy")

    SystemProxy = true
    
    c, err := GetSingBoxConfig(uuid, server)
    if err != nil {
        log.WithError(err).Error("Failed to get sing-box config")
        return err
    }
    
    // 创建基础 context
    ctx := context.Background()
    
    // 创建所有必要的注册器
    serviceRegistry := service.NewRegistry()
    ctx = service.ContextWith[service.Registry](ctx, serviceRegistry)
    
    // 创建注册器
    inboundRegistry := inbound.NewRegistry()
    outboundRegistry := outbound.NewRegistry()
    
    // 注册到 context
    ctx = service.ContextWith[adapter.InboundRegistry](ctx, inboundRegistry)
    ctx = service.ContextWith[adapter.OutboundRegistry](ctx, outboundRegistry)
    
    // 添加默认服务注册
    ctx = service.ContextWithDefaultRegistry(ctx)
    
    // 创建 box 实例
    instance, err := box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        log.WithError(err).Error("Failed to create sing-box instance")
        return E.Cause(err, "create client")
    }
    
    err = instance.Start()
    if err != nil {
        instance.Close()
        log.WithError(err).Error("Failed to start sing-box instance")
        return E.Cause(err, "start client")
    }
    
    log.Info("Proxy started successfully")
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
    }
    sysproxy.ClearSystemProxy()
    return nil
}