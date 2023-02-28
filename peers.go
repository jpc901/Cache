package Cache

import pb "Cache/gocachepb"

type PeerPicker interface {
	// PickPeer 于根据传入的 key 选择相应节点(选择相应的节点方法)
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	// Get 用于从对应 group 查找缓存值。
	Get(in *pb.Request, out *pb.Response) error
}
