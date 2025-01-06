package proxy

import (
    "context"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/adapter/outbound" 
    "github.com/sagernet/sing-box/adapter/inbound"
    "github.com/sagernet/sing-box/adapter/endpoint"
    "github.com/sagernet/sing-box/common/taskmonitor"
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

    log.WithField("config", c).Debug("Generated sing-box config")
    
    // 创建基础 context
    ctx := context.Background()

    // 创建注册器并注册服务
    serviceRegistry := service.NewRegistry()
    ctx = service.ContextWith[service.Registry](ctx, serviceRegistry)

    // 创建并注册 endpoint
    endpointRegistry := endpoint.NewRegistry()
    ctx = service.ContextWith[adapter.EndpointRegistry](ctx, endpointRegistry)
    
    // 创建并注册 inbound
    inboundRegistry := inbound.NewRegistry()
    ctx = service.ContextWith[adapter.InboundRegistry](ctx, inboundRegistry)
    
    // 创建并注册 outbound  
    outboundRegistry := outbound.NewRegistry()
    ctx = service.ContextWith[adapter.OutboundRegistry](ctx, outboundRegistry)

    // 添加默认服务注册
    ctx = service.ContextWithDefaultRegistry(ctx)

    // 创建任务监控
    monitor := taskmonitor.New(nil, 5)

    monitor.Start("Create sing-box instance")
    // 创建 box 实例
    instance, err := box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    monitor.Finish()

    if err != nil {
        log.WithError(err).Error("Failed to create sing-box instance")
        return E.Cause(err, "create client")
    }

    monitor.Start("Pre-start sing-box instance")
    // 先执行 PreStart
    err = instance.PreStart()
    monitor.Finish()

    if err != nil {
        instance.Close()
        log.WithError(err).Error("Failed to pre-start sing-box instance")
        return E.Cause(err, "pre-start client")
    }
    
    monitor.Start("Start sing-box instance")
    err = instance.Start()
    monitor.Finish()

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