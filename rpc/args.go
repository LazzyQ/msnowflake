package rpc

type NextIdsArgs struct {
	WorkerId int64 // snowflake 的workerId
	Num      int   // 批量获取id的数量
}
