package client

type Token struct {
	Namespace, Name string
}

type Config struct {
	Address string
	Token   Token
}
