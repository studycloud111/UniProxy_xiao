package proxy

import (
    "context"

    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/option"
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

// 实现必要的 Registry 接口
type emptyRegistry struct{}

func (r *emptyRegistry) Create(ctx context.Context, router adapter.Router, log service.Logger, tag string, options option.Options) (adapter.Endpoint, error) {
    return nil, nil
}

func (r *emptyRegistry) CreateInbound(ctx context.Context, router adapter.Router, log service.Logger, tag string, options option.Options) (adapter.Inbound, error) {
    return nil, nil
}

func (r *emptyRegistry) CreateOutbound(ctx context.Context, router adapter.Router, log service.Logger, tag string, options option.Options) (adapter.Outbound, error) {
    return nil, nil
}

func (r *emptyRegistry) CreateOptions(name string) (any, bool) {
    return nil, false
}

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
    
    // 创建空的 registry 实现
    registry := &emptyRegistry{}
    
    // 设置 registry
    ctx = service.ContextWithRegistry(ctx, service.NewRegistry())
    ctx = box.Context(ctx, registry, registry, registry)

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