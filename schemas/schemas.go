package schemas

import "http-multiplexer/lib/httpClientMultiplexer"

type ResponseGeneral struct {
	CountRequest int                                       `json:"count_request"`
	Responses    []httpClientMultiplexer.ResponseSingleUrl `json:"responses"`
}

type RequestBody struct {
	UrlList []string `json:"url_list"`
}

type Config struct {
	ServerPort         int    `env:"SERVER_PORT"`
	ServerPath         string `env:"SERVER_PATH"`
	ServerMaxCountReq  int    `env:"SERVER_MAX_COUNT_REQ"`
	HttpTimeout        int    `env:"HTTP_TIMEOUT"`
	HttpMaxCountURL    int    `env:"HTTP_MAX_COUNT_URL"`
	HttpMaxParallelReq int    `env:"HTTP_MAX_PARALLEL_REQ"`
}
