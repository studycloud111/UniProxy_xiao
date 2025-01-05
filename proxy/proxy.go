package proxy

import (
    "context"
    "google.golang.org/grpc"
    box "github.com/sagernet/sing-box"
    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/common/v2ray"
    "github.com/sagernet/sing-box/experimental"
    "github.com/sagernet/sing-box/option"
    slog "github.com/sagernet/sing-box/log"
    E "github.com/sagernet/sing/common/exceptions"
    "github.com/studycloud111/UniProxy_xiao/common/sysproxy"
    "github.com/studycloud111/UniProxy_xiao/v2b"
    v2rayproto "github.com/v2fly/v2ray-core/v5/app/stats/command"
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

// 统计服务实现
type statsService struct {
    stats map[string]int64
}

func (s *statsService) QueryStats(name string, reset bool) int64 {
    if value, exists := s.stats[name]; exists {
        if reset {
            delete(s.stats, name)
        }
        return value
    }
    return 0
}

func (s *statsService) GetStats() map[string]int64 {
    return s.stats
}

// V2Ray服务器实现
type v2rayServer struct {
    v2rayproto.UnimplementedStatsServiceServer
    logger     slog.Logger
    stats      *statsService
    grpcServer *grpc.Server
}

func (s *v2rayServer) Name() string {
    return "v2ray"
}

func (s *v2rayServer) Start(stage adapter.StartStage) error {
    if stage != adapter.StartStageMain {
        return nil
    }

    s.grpcServer = grpc.NewServer()
    v2rayproto.RegisterStatsServiceServer(s.grpcServer, s)
    
    return nil
}

func (s *v2rayServer) Close() error {
    if s.grpcServer != nil {
        s.grpcServer.GracefulStop()
    }
    return nil
}

func (s *v2rayServer) StatsService() adapter.StatsService {
    if s.stats == nil {
        s.stats = &statsService{
            stats: make(map[string]int64),
        }
    }
    return s.stats
}

// 实现 gRPC 统计服务接口
func (s *v2rayServer) GetStats(ctx context.Context, request *v2rayproto.GetStatsRequest) (*v2rayproto.GetStatsResponse, error) {
    if s.stats == nil {
        return &v2rayproto.GetStatsResponse{Stat: nil}, nil
    }
    
    value := s.stats.QueryStats(request.Name, request.Reset)
    return &v2rayproto.GetStatsResponse{
        Stat: &v2rayproto.Stat{
            Name:  request.Name,
            Value: value,
        },
    }, nil
}

func (s *v2rayServer) QueryStats(ctx context.Context, request *v2rayproto.QueryStatsRequest) (*v2rayproto.QueryStatsResponse, error) {
    if s.stats == nil {
        return &v2rayproto.QueryStatsResponse{Stat: nil}, nil
    }

    stats := s.stats.GetStats()
    var response v2rayproto.QueryStatsResponse
    
    for name, value := range stats {
        if request.Reset {
            delete(stats, name)
        }
        response.Stat = append(response.Stat, &v2rayproto.Stat{
            Name:  name,
            Value: value,
        })
    }
    
    return &response, nil
}

func init() {
    experimental.RegisterV2RayServerConstructor(func(logger slog.Logger, options option.V2RayAPIOptions) (adapter.V2RayServer, error) {
        return &v2rayServer{
            logger: logger,
            stats: &statsService{
                stats: make(map[string]int64),
            },
        }, nil
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