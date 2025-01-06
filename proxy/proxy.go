package proxy

import (
    "context"
    "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/adapter/outbound"
    "github.com/sagernet/sing-box/adapter/inbound"
    "github.com/sagernet/sing-box/adapter/endpoint"
    "github.com/sagernet/sing-box/log"
    E "github.com/sagernet/sing/common/exceptions"
    "github.com/sagernet/sing/service"
    "github.com/studycloud111/UniProxy_xiao/v2b"
    "github.com/studycloud111/UniProxy_xiao/common/sysproxy"
    log_service "github.com/sirupsen/logrus"
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

    log_service.WithFields(log_service.Fields{
        "tag": tag,
        "server_type": server.Type,
        "server_host": server.Host,
    }).Info("Starting proxy")

    SystemProxy = true
    
    c, err := GetSingBoxConfig(uuid, server)
    if err != nil {
        log_service.WithError(err).Error("Failed to get sing-box config")
        return err
    }
    
    // 创建基础 context
    ctx := context.Background()
    
    // 创建注册器
    endpointRegistry := endpoint.NewRegistry()
    inboundRegistry := inbound.NewRegistry()
    outboundRegistry := outbound.NewRegistry()
    
    // 注册到 context
    ctx = service.ContextWith[adapter.EndpointRegistry](ctx, endpointRegistry)
    ctx = service.ContextWith[adapter.InboundRegistry](ctx, inboundRegistry)
    ctx = service.ContextWith[adapter.OutboundRegistry](ctx, outboundRegistry)
    
    // 添加默认服务注册
    ctx = service.ContextWithDefaultRegistry(ctx)

    // 设置日志选项
    logFactory, err := log.NewFactory(log.Options{
        Context: ctx,
        Options: &log.Options{
            Level:  "debug",
            Output: "mixed",
        },
    })
    if err != nil {
        log_service.WithError(err).Error("Failed to create log factory")
        return E.Cause(err, "create log factory")
    }
    
    // 创建 box 实例
    instance, err := box.New(box.Options{
        Context: ctx,
        Options: c,
        PlatformLogWriter: logFactory.NewPlatformWriter(),
    })
    if err != nil {
        log_service.WithError(err).Error("Failed to create sing-box instance")
        return E.Cause(err, "create client")
    }
    
    err = instance.Start()
    if err != nil {
        instance.Close()
        log_service.WithError(err).Error("Failed to start sing-box instance")
        return E.Cause(err, "start client")
    }
    
    log_service.Info("Proxy started successfully")
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