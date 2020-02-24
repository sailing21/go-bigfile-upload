#golang实现文件分片上传     
需要redis     
实现了基本的文件秒传，断点续传，哈希检测，不重复合并等     
文件片段暂未删除，觉得需要删除的可以删除    
主要用于学习，用于生产环境可继续优化逻辑        
运行 
```
go run server.go fileDealer.go redisOp.go
```
