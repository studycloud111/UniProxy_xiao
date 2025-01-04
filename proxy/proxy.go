package proxy

import (
    "context"

    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/tree/dev-next/adapter/endpoint"
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
    
    // 创建 registry
    endpointRegistry := endpoint.NewRegistry()
    
    // 设置 registry
    ctx = box.Context(ctx, endpointRegistry, endpointRegistry, endpointRegistry)

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