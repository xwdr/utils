package utils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// 默认的Client, 如果配置了redis集群则默认redis.ClusterClient,否则默认redis.Client
var redisClient redis.Cmdable

// RedisSentinelConfig 哨兵配置
type RedisSentinelConfig struct {
	MasterName    string
	SentinelAddrs []string
	Password      string
	DB            int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
}

// NewSentinelnstance 初始化哨兵模式
func NewSentinelnstance(ctx context.Context, config *RedisSentinelConfig) *redis.Client {
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    config.MasterName,
		SentinelAddrs: config.SentinelAddrs,
		Password:      config.Password,
		DB:            config.DB,
		ReadTimeout:   config.ReadTimeout,
		WriteTimeout:  config.WriteTimeout,
		IdleTimeout:   config.IdleTimeout,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic("ping redis sentinel failed")
	}
	return redisClient
}

// RedisClusterlConfig Redis集群配
type RedisClusterlConfig struct {
	// 集群节点地址，理论上只要填一个可用的节点客户端就可以自动获取到集群的所有节点信息。但是最好多填一些节点以增加容灾能力，因为只填一个节点的话，如果这个节点出现了异常情况，则Go应用程序在启动过程中无法获取到集群信息。
	Addrs    []string `json:"addrs"`
	Password string   `json:"password"` // 密码

	// 只含读操作的命令的"节点选择策略"。默认都是false，即只能在主节点上执行。
	ReadOnly bool `json:"read_only"` // 置为true则允许在从节点上执行只含读操作的命令
	// 默认false。 置为true则ReadOnly自动置为true,表示在处理只读命令时，可以在一个slot对应的主节点和所有从节点中选取Ping()的响应时长最短的一个节点来读数据
	RouteByLatency bool `json:"route_by_latency"`
	// 默认false。置为true则ReadOnly自动置为true,表示在处理只读命令时，可以在一个slot对应的主节点和所有从节点中随机挑选一个节点来读数据
	RouteRandomly bool `json:"route_randomly"`
	//
	// // //用户可定制读取节点信息的函数，比如在非集群模式下可以从zookeeper读取。
	// // //但如果面向的是redis cluster集群，则客户端自动通过cluster slots命令从集群获取节点信息，不会用到这个函数。
	// ClusterSlots  func() ([]redis.ClusterSlot, error)
	//
	// // 钩子函数，当一个新节点创建时调用，传入的参数是新建的redis.Client
	// OnNewNode func(*redis.Client)
	//
	// // ------------------------------------------------------------------------------------------------------
	// // ClusterClient管理着一组redis.Client,下面的参数和非集群模式下的redis.Options参数一致，但默认值有差别。
	// // 初始化时，ClusterClient会把下列参数传递给每一个redis.Client
	//
	// // 钩子函数
	// // 仅当客户端执行命令需要从连接池获取连接时，如果连接池需要新建连接则会调用此钩子函数
	// OnConnect  func(conn *redis.Conn) error

	MaxRedirects int `json:"max_redirects"` // 当遇到网络错误或者MOVED/ASK重定向命令时，最多重试几次，默认8

	// 每一个redis.Client的连接池容量及闲置连接数量，而不是cluterClient总体的连接池大小。实际上没有总的连接池
	// 而是由各个redis.Client自行去实现和维护各自的连接池。
	PoolSize     int `json:"pool_size"`      // 连接池最大socket连接数，默认为5倍CPU数， 5 * runtime.NumCPU
	MinIdleConns int `json:"min_idle_conns"` // 在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

	// 命令执行失败时的重试策略
	MaxRetries      int           `json:"max_retries"`       // 命令执行失败时，最多重试多少次，默认为0即不重试
	MinRetryBackoff time.Duration `json:"min_retry_backoff"` // 每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
	MaxRetryBackoff time.Duration `json:"max_retry_backoff"` // 每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

	// 超时
	DialTimeout  time.Duration `json:"dial_timeout"`  // 连接建立超时时间，默认5秒。
	ReadTimeout  time.Duration `json:"read_timeout"`  // 读超时，默认3秒， -1表示取消读超时
	WriteTimeout time.Duration `json:"write_timeout"` // 写超时，默认等于读超时，-1表示取消读超时
	PoolTimeout  time.Duration `json:"pool_timeout"`  // 当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

	// 闲置连接检查包括IdleTimeout，MaxConnAge
	IdleCheckFrequency time.Duration `json:"idle_check_frequency"` // 闲置连接检查的周期，无默认值，由ClusterClient统一对所管理的redis.Client进行闲置连接检查。初始化时传递-1给redis.Client表示redis.Client自己不用做周期性检查，只在客户端获取连接时对闲置连接进行处理。
	IdleTimeout        time.Duration `json:"idle_timeout"`         // 闲置超时，默认5分钟，-1表示取消闲置超时检查
	MaxConnAge         time.Duration `json:"max_conn_age"`         // 连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接
}

// NewClusterlnstance 初始化集群
func NewClusterlnstance(ctx context.Context, config *RedisClusterlConfig) *redis.ClusterClient {
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		// -------------------------------------------------------------------------------------------
		// 集群相关的参数

		// 集群节点地址，理论上只要填一个可用的节点客户端就可以自动获取到集群的所有节点信息。但是最好多填一些节点以增加容灾能力，因为只填一个节点的话，如果这个节点出现了异常情况，则Go应用程序在启动过程中无法获取到集群信息。
		Addrs: config.Addrs,
		// 账号：  root  密码： redis123456
		Password:     config.Password,
		MaxRedirects: config.MaxRedirects, // 当遇到网络错误或者MOVED/ASK重定向命令时，最多重试几次，默认8

		// 只含读操作的命令的"节点选择策略"。默认都是false，即只能在主节点上执行。
		ReadOnly: config.ReadOnly, // 置为true则允许在从节点上执行只含读操作的命令
		// 默认false。 置为true则ReadOnly自动置为true,表示在处理只读命令时，可以在一个slot对应的主节点和所有从节点中选取Ping()的响应时长最短的一个节点来读数据
		RouteByLatency: config.RouteByLatency,
		// 默认false。置为true则ReadOnly自动置为true,表示在处理只读命令时，可以在一个slot对应的主节点和所有从节点中随机挑选一个节点来读数据
		RouteRandomly: config.RouteRandomly,

		// //用户可定制读取节点信息的函数，比如在非集群模式下可以从zookeeper读取。
		// //但如果面向的是redis cluster集群，则客户端自动通过cluster slots命令从集群获取节点信息，不会用到这个函数。
		// ClusterSlots: func() ([]redis.ClusterSlot, error) {

		// },

		// ------------------------------------------------------------------------------------------------------
		// ClusterClient管理着一组redis.Client,下面的参数和非集群模式下的redis.Options参数一致，但默认值有差别。
		// 初始化时，ClusterClient会把下列参数传递给每一个redis.Client

		// 钩子函数
		// 仅当客户端执行命令需要从连接池获取连接时，如果连接池需要新建连接则会调用此钩子函数
		OnConnect: func(ctx context.Context, conn *redis.Conn) error {
			return nil
		},

		// 每一个redis.Client的连接池容量及闲置连接数量，而不是cluterClient总体的连接池大小。实际上没有总的连接池
		// 而是由各个redis.Client自行去实现和维护各自的连接池。
		PoolSize:     config.PoolSize,     // 连接池最大socket连接数，默认为5倍CPU数， 5 * runtime.NumCPU
		MinIdleConns: config.MinIdleConns, // 在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

		// 命令执行失败时的重试策略
		MaxRetries:      config.MaxRetries,      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: config.MinRetryBackoff, // 每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: config.MaxRetryBackoff, // 每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		// 超时
		DialTimeout:  config.DialTimeout,  // 连接建立超时时间，默认5秒。
		ReadTimeout:  config.ReadTimeout,  // 读超时，默认3秒， -1表示取消读超时
		WriteTimeout: config.WriteTimeout, // 写超时，默认等于读超时，-1表示取消读超时
		PoolTimeout:  config.PoolTimeout,  // 当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		// 闲置连接检查包括IdleTimeout，MaxConnAge
		IdleCheckFrequency: config.IdleCheckFrequency, // 闲置连接检查的周期，无默认值，由ClusterClient统一对所管理的redis.Client进行闲置连接检查。初始化时传递-1给redis.Client表示redis.Client自己不用做周期性检查，只在客户端获取连接时对闲置连接进行处理。
		IdleTimeout:        config.IdleTimeout,        // 闲置超时，默认5分钟，-1表示取消闲置超时检查
		MaxConnAge:         config.MaxConnAge,         // 连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic("ping redis cluster failed")
	}
	return redisClient
}
