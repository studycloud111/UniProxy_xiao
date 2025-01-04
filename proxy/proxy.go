package proxy

import (
    "context"
    "net/netip"  // 添加这个导入

    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/option"  // 添加这个导入
    "github.com/sagernet/sing/service"     // 添加这个导入
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

    // 创建带 registry 的 context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 创建一个新的 registry
    registry := service.NewRegistry(ctx)

    // 将 registry 添加到 context
    ctx = service.ContextWithRegistry(ctx, registry)

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