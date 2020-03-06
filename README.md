# 使用方法



+ 可直接用仓库里的  binary
+ 使用 client 的时候 要配置 client.conf

```
LogLevel = "debug"
# server 的 ip + port
Addr     = "127.0.0.1:8080"
# 本机 hostname
Hostname = "gpu0"
# 是否自动监控
AutoKill = true
# 每 Interval 秒检测一次
Interval = 3
# 最低 gpu 使用率  
Low      =  50
# 当一个进程 最低 gpu 使用率 连续 Times 次低于 Low 那么会自动 Kill 掉这个进程
Times    = 60
```

