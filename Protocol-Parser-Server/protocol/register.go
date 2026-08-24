// 初始化注册入口
package protocol

import (
	"protocol-parser-server/parser/core"
	"protocol-parser-server/protocol/jt808"
	"protocol-parser-server/protocol/p2929"
	"protocol-parser-server/protocol/vdf"
)

func RegisterProtocols() {

	core.Register(p2929.New())
	core.Register(jt808.New())
	core.Register(vdf.New())

}
