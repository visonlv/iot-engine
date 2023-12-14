package group

import pb "github.com/visonlv/iot-engine/group/proto"

func Start(param *Param) (*Client, *pb.CategoryHeartBeatResp, error) {
	c := newClient(param)
	data, err := c.get()
	go c.mainloop()
	return c, data, err
}
