module rtr/rtrtcp

go 1.15

require (
	github.com/astaxie/beego v1.12.3
	github.com/cpusoft/goutil v1.0.18
	model v0.0.0-00010101000000-000000000000
	rtr/db v0.0.0-00010101000000-000000000000
	rtr/model v0.0.0-00010101000000-000000000000
	uint128 v0.0.0-00010101000000-000000000000
)

replace (
	model => ../../model
	rtr/db => ../../rtr/db
	rtr/model => ../../rtr/model
	uint128 => ../../uint128
)
