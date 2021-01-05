# 跟着geektutu学习golang 之 缓存篇

## 缓存淘汰算法: FIFO/LFU/LRU

1. FIFO (First In First Out)

先进先出，也就是淘汰缓存中最老(最早添加)的记录。FIFO 认为，最早添加的记录，
其不再被使用的可能性比刚添加的可能性大。这种算法的实现也非常简单，
创建一个队列，新增记录添加到队尾，每次内存不够时，淘汰队首。但是很多场景下，
部分记录虽然是最早添加但也最常被访问，而不得不因为呆的时间太长而被淘汰。
这类数据会被频繁地添加进缓存，又被淘汰出去，导致缓存命中率降低。

2. LFU(Least Frequently Used)

 最少使用，也就是淘汰缓存中访问频率最低的记录。LFU 认为，如果数据过去被访问
 多次，那么将来被访问的频率也更高。LFU 的实现需要维护一个按照访问次数排序的
 队列，每次访问，访问次数加1，队列重新排序，淘汰时选择访问次数最少的即可。
 LFU 算法的命中率是比较高的，但缺点也非常明显，维护每个记录的访问次数，对内
 存的消耗是很高的；另外，如果数据的访问模式发生变化，LFU 需要较长的时间去适
 应，也就是说 LFU 算法受历史数据的影响比较大。例如某个数据历史上访问次数奇
 高，但在某个时间点之后几乎不再被访问，但因为历史访问次数过高，而迟迟不能被
 淘汰。

3. LRU(Least Recently Used)

最近最少使用，相对于仅考虑时间因素的 FIFO 和仅考虑访问频率的 LFU，LRU 算法
可以认为是相对平衡的一种淘汰算法。LRU 认为，如果数据最近被访问过，那么将来
被访问的概率也会更高。LRU 算法的实现非常简单，维护一个队列，如果某条记录被
访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。

## 如何控制并发时多个缓存不被同时调用
多个协程(goroutine)同时读写同一个变量，在并发度较高的情况下，会发生冲突。
确保一次只有一个协程(goroutine)可以访问该变量以避免冲突，这称之为互斥.

>sync.Mutex 是一个互斥锁，可以由不同的协程加锁和解锁。

sync.Mutex 是 Go 语言标准库提供的一个互斥锁，当一个协程(goroutine)获得了
这个锁的拥有权后，其它请求锁的协程(goroutine) 就会阻塞在 Lock() 方法的调
用上，直到调用 Unlock() 后锁被释放。

## 缓存雪崩/缓存击穿/缓存穿透

>缓存雪崩：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。缓存雪崩通常因为缓存服务器宕机、缓存的 key 设置了相同的过期时间等引起。

>缓存击穿：一个存在的key，在缓存过期的一刻，同时有大量的请求，这些请求都会击穿到 DB ，造成瞬时DB请求量大、压力骤增。

>缓存穿透：查询一个不存在的数据，因为不存在则不会写到缓存中，所以每次都会去请求 DB，如果瞬间流量过大，穿透到 DB，导致宕机。

## 防止缓存雪崩/缓存击穿/缓存穿透的策略
> singleflight 