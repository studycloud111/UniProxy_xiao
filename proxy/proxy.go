package proxy

import (
    "context"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/experimental"
    "github.com/sagernet/sing-box/option"
    slog "github.com/sagernet/sing-box/log"
    E "github.com/sagernet/sing/common/exceptions"
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

// 实现一个简单的 V2Ray 服务器
type v2rayServer struct {
    logger slog.Logger
}

func (s *v2rayServer) Name() string {
    return "v2ray"
}

func (s *v2rayServer) Start(stage adapter.StartStage) error {
    return nil
}

func (s *v2rayServer) Close() error {
    return nil
}

func (s *v2rayServer) StatsService() adapter.V2RayStatsService {
    return nil
}

func init() {
    experimental.RegisterV2RayServerConstructor(func(logger slog.Logger, options option.V2RayAPIOptions) (adapter.V2RayServer, error) {
        return &v2rayServer{logger: logger}, nil
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
    
    ctx := context.Background()
    
    client, err = box.New(box.Options{
        Context: ctx,
        Options: c,
    })
    if err != nil {
        return E.Cause(err, "create client")
    }
    
    err = client.Start()
    if err != nil {
        return E.Cause(err, "start client")
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