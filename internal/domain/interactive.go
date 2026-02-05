package domain

type Interactive struct {
	// 阅读点赞收藏数量
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	// 我有没有点赞和收藏
	Liked     bool
	Collected bool
}
