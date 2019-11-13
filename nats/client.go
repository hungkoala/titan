package nats

type Client struct {
	Addr string
}

func (srv *Client) Request(rq *Request) (*Response, error) {
	c, err := NewConnection(srv.Addr)
	if err != nil {
		return nil, err
	}
	defer func(c *Connection) {
		c.Conn.Close()
	}(c)

	return c.SendRequest(rq)
}

func NewClient(addr string) *Client {
	return &Client{addr}
}
