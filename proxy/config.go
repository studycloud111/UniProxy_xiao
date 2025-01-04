package proxy

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/hex"
    "errors"
    "fmt"
    "net/netip"
    "net/url"
    "os"
    "path"
    "strconv"
    "strings"

    C "github.com/sagernet/sing-box/constant"
    "github.com/sagernet/sing-box/option"
    "github.com/sagernet/sing/common/json/badoption"
    "github.com/studycloud111/UniProxy_xiao/common/file"
    "github.com/studycloud111/UniProxy_xiao/geo"
    "github.com/studycloud111/UniProxy_xiao/v2b"
)

func GetSingBoxConfig(uuid string, server *v2b.ServerInfo) (option.Options, error) {
    var in option.Inbound
    if TunMode {
        in.Type = "tun"
        tunOptions := option.TunInboundOptions{
            MTU:         9000,
            AutoRoute:   true,
            StrictRoute: true,
            Stack:       "gvisor",
        }
        
        prefix4, _ := netip.ParsePrefix("172.19.0.1/30")
        prefix6, _ := netip.ParsePrefix("fdfe:dcba:9876::1/126")
        tunOptions.Address = []netip.Prefix{prefix4, prefix6}
        
        route4_1, _ := netip.ParsePrefix("0.0.0.0/1")
        route4_2, _ := netip.ParsePrefix("128.0.0.0/1")
        route6_1, _ := netip.ParsePrefix("::/1")
        route6_2, _ := netip.ParsePrefix("8000::/1")
        tunOptions.RouteAddress = []netip.Prefix{route4_1, route4_2, route6_1, route6_2}
        
        in.Options = tunOptions
    } else {
        in.Type = "mixed"
        addr := netip.MustParseAddr("127.0.0.1")
        
        mixedOptions := option.HTTPMixedInboundOptions{
            ListenOptions: option.ListenOptions{
                Listen:     (*badoption.Addr)(&addr),
                ListenPort: uint16(InPort),
            },
        }
        
        in.Options = mixedOptions
    }

    var serverPort uint16
    if strings.Contains(server.Port, "-") {
        ports := strings.Split(server.Port, "-")
        if len(ports) > 0 {
            port, err := strconv.ParseUint(ports[0], 10, 16)
            if err != nil {
                return option.Options{}, fmt.Errorf("invalid port number: %s", err)
            }
            serverPort = uint16(port)
        }
    } else {
        port, err := strconv.ParseUint(server.Port, 10, 16)
        if err != nil {
            return option.Options{}, fmt.Errorf("invalid port number: %s", err)
        }
        serverPort = uint16(port)
    }

    var out option.Outbound
    out.Tag = "proxy"
    so := option.ServerOptions{
        Server:     server.Host,
        ServerPort: serverPort,
    }
    
    switch server.Type {
    case "vmess":
        out.Type = "vmess"
        vmessOptions := option.VMessOutboundOptions{
            ServerOptions:       so,
            UUID:               uuid,
            Security:           "auto",
            AuthenticatedLength: true,
        }

        if server.Network != "" && server.Network != "tcp" {
            transport := &option.V2RayTransportOptions{
                Type: server.Network,
            }
            
            switch server.Network {
            switch server.Network {
			case "ws":
			    u, err := url.Parse(server.NetworkSettings.Path)
			    if err != nil {
			        return option.Options{}, err
			    }
			    ed, _ := strconv.Atoi(u.Query().Get("ed"))
			    transport.WebsocketOptions = option.V2RayWebsocketOptions{  // 移除 &
			        Path: u.Path,
			        MaxEarlyData: uint32(ed),
			        EarlyDataHeaderName: "Sec-WebSocket-Protocol",
			    }
			case "grpc":
			    transport.GRPCOptions = option.V2RayGRPCOptions{  // 移除 &
			        ServiceName: server.ServerName,
			    }
			}
            
            vmessOptions.Transport = transport
        }

        if server.Tls == 1 {
            vmessOptions.TLS = &option.OutboundTLSOptions{
                Enabled:    true,
                ServerName: server.ServerName,
                Insecure:   server.TlsSettings.AllowInsecure != "0",
            }
        }
        
        out.Options = vmessOptions

    case "vless":
        out.Type = "vless"
        vlessOptions := option.VLESSOutboundOptions{
            ServerOptions: so,
            UUID:         uuid,
            Flow:         server.Flow,
        }

        if server.Network != "" && server.Network != "tcp" {
            transport := &option.V2RayTransportOptions{
                Type: server.Network,
            }
            
            switch server.Network {
            case "ws":
                u, err := url.Parse(server.NetworkSettings.Path)
                if err != nil {
                    return option.Options{}, err
                }
                ed, _ := strconv.Atoi(u.Query().Get("ed"))
                transport.WebsocketOptions = &option.V2RayWebsocketOptions{
                    Path: u.Path,
                    MaxEarlyData: uint32(ed),
                    EarlyDataHeaderName: "Sec-WebSocket-Protocol",
                }
            case "grpc":
                transport.GRPCOptions = &option.V2RayGRPCOptions{
                    ServiceName: server.ServerName,
                }
            }
            
            vlessOptions.Transport = transport
        }

        if server.Tls >= 1 {
            tlsOptions := &option.OutboundTLSOptions{
                Enabled:    true,
                ServerName: server.ServerName,
                Insecure:   server.TlsSettings.AllowInsecure != "0",
            }
            
            if server.Tls == 2 {
                tlsOptions.UTLS = &option.OutboundUTLSOptions{
                    Enabled:     true,
                    Fingerprint: server.TlsSettings.Fingerprint,
                }
                tlsOptions.Reality = &option.OutboundRealityOptions{
                    Enabled:   true,
                    ShortID:   server.TlsSettings.ShortId,
                    PublicKey: server.TlsSettings.PublicKey,
                }
            }
            vlessOptions.TLS = tlsOptions
        }

        out.Options = vlessOptions

    case "shadowsocks":
        out.Type = "shadowsocks"
        var keyLength int
        switch server.Cipher {
        case "2022-blake3-aes-128-gcm":
            keyLength = 16
        case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
            keyLength = 32
        }
        
        var pw string
        if keyLength != 0 {
            createdAtString := strconv.Itoa(server.CreatedAt)
            hash := md5.Sum([]byte(createdAtString))
            pw = base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(hash[:])[:keyLength])) + ":" + base64.StdEncoding.EncodeToString([]byte(uuid[:keyLength]))
        } else {
            pw = uuid
        }

        out.Options = option.ShadowsocksOutboundOptions{
            ServerOptions: so,
            Password:     pw,
            Method:       server.Cipher,
        }

    case "trojan":
        out.Type = "trojan"
        trojanOptions := option.TrojanOutboundOptions{
            ServerOptions: so,
            Password:     uuid,
        }

        if server.Network != "" && server.Network != "tcp" {
            transport := &option.V2RayTransportOptions{
                Type: server.Network,
            }
            
            switch server.Network {
            case "ws":
                u, err := url.Parse(server.NetworkSettings.Path)
                if err != nil {
                    return option.Options{}, err
                }
                ed, _ := strconv.Atoi(u.Query().Get("ed"))
                transport.WebsocketOptions = &option.V2RayWebsocketOptions{
                    Path: u.Path,
                    MaxEarlyData: uint32(ed),
                    EarlyDataHeaderName: "Sec-WebSocket-Protocol",
                }
            case "grpc":
                transport.GRPCOptions = &option.V2RayGRPCOptions{
                    ServiceName: server.ServerName,
                }
            }
            
            trojanOptions.Transport = transport
        }

        if server.Tls == 1 {
            trojanOptions.TLS = &option.OutboundTLSOptions{
                Enabled:    true,
                ServerName: server.ServerName,
                Insecure:   server.Allow_Insecure == 1,
            }
        }

        out.Options = trojanOptions

    case "hysteria":
        if server.HysteriaVersion == 2 {
            out.Type = "hysteria2"
            
            var obfs *option.Hysteria2Obfs
            if server.Hy2Obfs != "" && server.Hy2ObfsPassword != "" {
                obfs = &option.Hysteria2Obfs{
                    Type:     server.Hy2Obfs,
                    Password: server.Hy2ObfsPassword,
                }
            } else if server.Hy2Obfs != "" {
                obfs = &option.Hysteria2Obfs{
                    Type:     "salamander",
                    Password: server.Hy2Obfs,
                }
            }
            
            hy2Options := option.Hysteria2OutboundOptions{
                ServerOptions: option.ServerOptions{
                    Server: server.Host,
                },
                Obfs:     obfs,
                Password: uuid,
                UpMbps:   server.UpMbps,
                DownMbps: server.DownMbps,
            }
            
            // TLS 配置
            hy2Options.TLS = &option.OutboundTLSOptions{
                Enabled:    true,
                Insecure:   server.AllowInsecure == 1,
                ServerName: server.ServerName,
            }

            // 根据 Mport 的格式决定端口配置
            if strings.Contains(server.Mport, "-") {
                hy2Options.ServerPorts = badoption.Listable[string]{server.Mport}
            } else {
                port, _ := strconv.ParseUint(server.Mport, 10, 16)
                hy2Options.ServerOptions.ServerPort = uint16(port)
            }

            out.Options = hy2Options

        } else {
            out.Type = "hysteria"
            
            hy1Options := option.HysteriaOutboundOptions{
                ServerOptions: option.ServerOptions{
                    Server: server.Host,
                },
                UpMbps:     server.UpMbps,
                DownMbps:   server.DownMbps,
                Obfs:       server.ServerKey,
                AuthString: uuid,
            }

            hy1Options.TLS = &option.OutboundTLSOptions{
                Enabled:    true,
                Insecure:   server.AllowInsecure == 1,
                ServerName: server.ServerName,
            }

            port, _ := strconv.ParseUint(server.Mport, 10, 16)
            hy1Options.ServerOptions.ServerPort = uint16(port)

            out.Options = hy1Options
        }

    default:
        return option.Options{}, errors.New("server type is unknown")
    }

    r, err := getRules(GlobalMode)
    if err != nil {
        return option.Options{}, fmt.Errorf("get rules error: %s", err)
    }

    return option.Options{
        Log: &option.LogOptions{
            Level: "debug",  // 设置为debug级别以便排查问题
        },
        Inbounds: []option.Inbound{
            in,
        },
        Outbounds: []option.Outbound{
            out,
            {
                Tag:  "direct",
                Type: "direct",
            },
        },
        Route: r,
    }, nil
}

func getRules(global bool) (*option.RouteOptions, error) {
    if global {
        return &option.RouteOptions{
            AutoDetectInterface: true,
        }, nil
    }

    err := checkRes(DataPath)
    if err != nil {
        return nil, fmt.Errorf("check res err: %s", err)
    }

    return &option.RouteOptions{
        GeoIP: &option.GeoIPOptions{
            DownloadURL: ResUrl + "/geoip.db",
            Path:        path.Join(DataPath, "geoip.dat"),
        },
        Geosite: &option.GeositeOptions{
            DownloadURL: ResUrl + "/geosite.db",
            Path:        path.Join(DataPath, "geosite.dat"),
        },
        Rules: []option.Rule{
            {
                Type: C.RuleTypeDefault,
                DefaultOptions: option.DefaultRule{
                    RawDefaultRule: option.RawDefaultRule{
                        GeoIP:   badoption.Listable[string]{"cn", "private"},
                        Geosite: badoption.Listable[string]{"cn"},
                    },
                    RuleAction: option.RuleAction{
                        Action: "direct",
                    },
                },
            },
        },
    }, nil
}

func checkRes(p string) error {
    if !file.IsExist(path.Join(p, "geoip.dat")) {
        f, err := os.OpenFile(path.Join(p, "geoip.dat"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
        if err != nil {
            return err
        }
        defer f.Close()
        _, err = f.Write(geo.Ip)
        if err != nil {
            return err
        }
    }
    if !file.IsExist(path.Join(p, "geosite.dat")) {
        f, err := os.OpenFile(path.Join(p, "geosite.dat"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
        if err != nil {
            return err
        }
        defer f.Close()
        _, err = f.Write(geo.Site)
        if err != nil {
            return err
        }
    }
    return nil
}