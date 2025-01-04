package proxy

import (
    "context"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/experimental"
    "github.com/sagernet/sing-box/option"
    slog "github.com/sagernet/sing-box/log"  // 使用 sing-box 的 log 包
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

func init() {
    // 修改为正确的 logger 类型
    experimental.RegisterV2RayServerConstructor(func(logger slog.Logger, options option.V2RayAPIOptions) (adapter.V2RayServer, error) {
        return nil, nil
    })
}

func StartProxy(tag string, uuid string, server *v2b.ServerInfo) error {
    if Running {
        StopProxy()
    }
    SystemProxy = true
    
    c, err := GetSingBoxConfig(uuid, server)
    if err != nil {
        return err
    }
    
    // 只使用基础 context
    ctx := context.Background()
    
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