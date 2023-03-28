package server

type ServerOptions struct {
	Listen          string
	Hu60wap6APIURL  string
	DisabledActions []string
}

type CanalOptions struct {
	CanalHost              string
	CanalPort              int
	CanalClientDestination string
}
