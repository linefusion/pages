package config

type ServerConfig struct {
	Name    string       `hcl:"name,label"`
	Listen  ListenBlock  `hcl:"listen,block"`
	Options OptionsBlock `hcl:"options,optional"`
	Pages   PageBlocks   `hcl:"pages,block"`
}

type ListenBlock struct {
	Bind string `hcl:"bind,optional"`
	Port int    `hcl:"port"`
}

type OptionsBlock struct {
	ReadTimeout                        *int  `hcl:"read_timeout,optional"`
	WriteTimeout                       *int  `hcl:"write_timeout,optional"`
	IdleTimeout                        *int  `hcl:"idle_timeout,optional"`
	Concurrency                        *int  `hcl:"concurrency,optional"`
	DisableKeepalive                   *bool `hcl:"disable_keepalive,optional"`
	ReadBufferSize                     *int  `hcl:"read_buffer_size,optional"`
	WriteBufferSize                    *int  `hcl:"write_buffer_size,optional"`
	MaxConnsPerIP                      *int  `hcl:"max_conns_per_ip,optional"`
	MaxRequestsPerConn                 *int  `hcl:"max_requests_per_conn,optional"`
	TCPKeepalive                       *bool `hcl:"tcp_keepalive,optional"`
	TCPKeepalivePeriod                 *int  `hcl:"tcp_keepalive_period,optional"`
	MaxRequestBodySize                 *int  `hcl:"max_request_body_size,optional"`
	ReduceMemoryUsage                  *bool `hcl:"reduce_memory_usage,optional"`
	GetOnly                            *bool `hcl:"get_only,optional"`
	DisablePreParseMultipartForm       *bool `hcl:"disable_pre_parse_multipart_form,optional"`
	LogAllErrors                       *bool `hcl:"log_all_errors,optional"`
	SecureErrorLogMessage              *bool `hcl:"secure_error_log_message,optional"`
	DisableHeaderNamesNormalizing      *bool `hcl:"disable_header_names_normalizing,optional"`
	SleepWhenConcurrencyLimitsExceeded *int  `hcl:"sleep_when_concurrency_limits_exceeded,optional"`
	NoDefaultServerHeader              *bool `hcl:"no_default_server_header,optional"`
	NoDefaultDate                      *bool `hcl:"no_default_date,optional"`
	NoDefaultContentType               *bool `hcl:"no_default_content_type,optional"`
	KeepHijackedConns                  *bool `hcl:"keep_hijacked_conns,optional"`
	CloseOnShutdown                    *bool `hcl:"close_on_shutdown,optional"`
	StreamRequestBody                  *bool `hcl:"stream_request_body,optional"`
	//MaxKeepaliveDuration               *int  `hcl:"max_keepalive_duration,optional"`
}
